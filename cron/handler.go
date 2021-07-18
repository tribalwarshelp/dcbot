package cron

import (
	"context"
	"fmt"
	"github.com/Kichiyaki/appmode"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/pkg/errors"

	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/util/twutil"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/model"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"
)

type handler struct {
	lastEnnoblementAt map[string]time.Time
	serverRepo        server.Repository
	observationRepo   observation.Repository
	groupRepo         group.Repository
	discord           *discord.Session
	api               *sdk.SDK
	status            string
}

func (h *handler) loadEnnoblements(servers []string) (map[string]ennoblements, error) {
	m := make(map[string]ennoblements)

	if len(servers) == 0 {
		return m, nil
	}

	query := ""

	for _, s := range servers {
		lastEnnoblementAt, ok := h.lastEnnoblementAt[s]
		if !ok {
			lastEnnoblementAt = time.Now().Add(-1 * time.Minute)
			h.lastEnnoblementAt[s] = lastEnnoblementAt
		}
		if appmode.Equals(appmode.DevelopmentMode) {
			lastEnnoblementAt = time.Now().Add(-1 * time.Hour * 2)
		}
		lastEnnoblementAtJSON, err := lastEnnoblementAt.MarshalJSON()
		if err != nil {
			continue
		}
		query += fmt.Sprintf(`
			%s: ennoblements(server: "%s", filter: { ennobledAtGT: %s }) {
				items {
					%s
					ennobledAt
				}
			}
		`, s,
			s,
			string(lastEnnoblementAtJSON),
			sdk.EnnoblementInclude{
				NewOwner: true,
				Village:  true,
				NewOwnerInclude: sdk.PlayerInclude{
					Tribe: true,
				},
				OldOwner: true,
				OldOwnerInclude: sdk.PlayerInclude{
					Tribe: true,
				},
			}.String())
	}

	resp := make(map[string]*sdk.EnnoblementList)
	if err := h.api.Post(fmt.Sprintf(`query { %s }`, query), &resp); err != nil {
		return m, errors.Wrap(err, "loadEnnoblements")
	}

	for s, singleServerResp := range resp {
		if singleServerResp == nil {
			continue
		}
		m[s] = singleServerResp.Items
		lastEnnoblement := m[s].getLastEnnoblement()
		if lastEnnoblement != nil {
			h.lastEnnoblementAt[s] = lastEnnoblement.EnnobledAt
		}
	}

	return m, nil
}

func (h *handler) loadVersions(servers []string) ([]*twmodel.Version, error) {
	var versionCodes []twmodel.VersionCode
	cache := make(map[twmodel.VersionCode]bool)
	for _, s := range servers {
		versionCode := twmodel.VersionCodeFromServerKey(s)
		if versionCode.IsValid() && !cache[versionCode] {
			cache[versionCode] = true
			versionCodes = append(versionCodes, versionCode)
		}
	}

	if len(versionCodes) == 0 {
		return []*twmodel.Version{}, nil
	}

	versionList, err := h.api.Version.Browse(0, 0, []string{"code ASC"}, &twmodel.VersionFilter{
		Code: versionCodes,
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load versions")
	}

	return versionList.Items, nil
}

func (h *handler) checkEnnoblements() {
	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Errorln("checkEnnoblements:", err.Error())
		return
	}
	log.
		WithField("servers", servers).
		Info("checkEnnoblements: servers have been loaded")

	groups, total, err := h.groupRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Errorln("checkEnnoblements:", err.Error())
		return
	}
	log.
		WithField("numberOfGroups", total).
		Info("checkEnnoblements: groups have been loaded")

	versions, err := h.loadVersions(servers)
	if err != nil {
		log.Errorln("checkEnnoblements:", err)
		return
	}
	log.
		WithField("numberOfVersions", len(versions)).
		Info("checkEnnoblements: versions have been loaded")

	ennoblementsByServerKey, err := h.loadEnnoblements(servers)
	if err != nil {
		log.Errorln("checkEnnoblements:", err)
	}
	log.Info("checkEnnoblements: ennoblements have been loaded")

	for _, g := range groups {
		if g.ConqueredVillagesChannelID == "" && g.LostVillagesChannelID == "" {
			continue
		}
		localizer := message.NewLocalizer(g.Server.Lang)
		lostVillagesMsg := &discord.MessageEmbedFieldBuilder{}
		conqueredVillagesMsg := &discord.MessageEmbedFieldBuilder{}
		for _, obs := range g.Observations {
			enblmnts, ok := ennoblementsByServerKey[obs.Server]
			version := twutil.FindVersionByCode(versions, twmodel.VersionCodeFromServerKey(obs.Server))
			if ok && version != nil && version.Host != "" {
				if g.LostVillagesChannelID != "" {
					for _, ennoblement := range enblmnts.getLostVillagesByTribe(obs.TribeID) {
						if !twutil.IsPlayerTribeNil(ennoblement.NewOwner) &&
							g.Observations.Contains(obs.Server, ennoblement.NewOwner.Tribe.ID) {
							continue
						}
						newMsgDataConfig := newMessageConfig{
							host:        version.Host,
							server:      obs.Server,
							ennoblement: ennoblement,
							t:           messageTypeLost,
							localizer:   localizer,
						}
						lostVillagesMsg.Append(newMessage(newMsgDataConfig).String())
					}
				}

				if g.ConqueredVillagesChannelID != "" {
					for _, ennoblement := range enblmnts.getConqueredVillagesByTribe(obs.TribeID, g.ShowInternals) {
						isInTheSameGroup := !twutil.IsPlayerTribeNil(ennoblement.OldOwner) &&
							g.Observations.Contains(obs.Server, ennoblement.OldOwner.Tribe.ID)
						if (!g.ShowInternals && isInTheSameGroup) ||
							(!g.ShowEnnobledBarbarians && isBarbarian(ennoblement.OldOwner)) {
							continue
						}

						newMsgDataConfig := newMessageConfig{
							host:        version.Host,
							server:      obs.Server,
							ennoblement: ennoblement,
							t:           messageTypeConquer,
							localizer:   localizer,
						}
						conqueredVillagesMsg.Append(newMessage(newMsgDataConfig).String())
					}
				}
			}
		}

		timestamp := time.Now().Format(time.RFC3339)
		if g.ConqueredVillagesChannelID != "" && !conqueredVillagesMsg.IsEmpty() {
			title := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CronConqueredVillagesTitle,
				DefaultMessage: message.FallbackMsg(message.CronConqueredVillagesTitle,
					"Conquered villages"),
			})
			go h.discord.SendEmbed(g.ConqueredVillagesChannelID,
				discord.
					NewEmbed().
					SetTitle(title).
					SetColor(colorConqueredVillages).
					SetFields(conqueredVillagesMsg.ToMessageEmbedFields()).
					SetTimestamp(timestamp))
		}

		if g.LostVillagesChannelID != "" && !lostVillagesMsg.IsEmpty() {
			title := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CronLostVillagesTitle,
				DefaultMessage: message.FallbackMsg(message.CronLostVillagesTitle,
					"Lost villages"),
			})
			go h.discord.SendEmbed(g.LostVillagesChannelID,
				discord.
					NewEmbed().
					SetTitle(title).
					SetColor(colorLostVillages).
					SetFields(lostVillagesMsg.ToMessageEmbedFields()).
					SetTimestamp(timestamp))
		}
	}
}

func (h *handler) checkBotServers() {
	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Error("checkBotServers: " + err.Error())
		return
	}
	log.
		WithField("numberOfServers", total).
		Info("checkBotServers: loaded servers")

	var idsToDelete []string
	for _, s := range servers {
		if isGuildMember, _ := h.discord.IsGuildMember(s.ID); !isGuildMember {
			idsToDelete = append(idsToDelete, s.ID)
		}
	}

	if len(idsToDelete) > 0 {
		deleted, err := h.serverRepo.Delete(context.Background(), &model.ServerFilter{
			ID: idsToDelete,
		})
		if err != nil {
			log.Error("checkBotServers: " + err.Error())
		} else {
			log.
				WithField("numberOfDeletedServers", len(deleted)).
				Info("checkBotServers: some of the servers have been deleted")
		}
	}
}

func (h *handler) deleteClosedTWServers() {
	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Error("deleteClosedTWServers: " + err.Error())
		return
	}
	log.
		WithField("servers", servers).
		Info("deleteClosedTWServers: loaded servers")

	list, err := h.api.Server.Browse(0, 0, []string{"key ASC"}, &twmodel.ServerFilter{
		Key:    servers,
		Status: []twmodel.ServerStatus{twmodel.ServerStatusClosed},
	}, nil)
	if err != nil {
		log.Errorln("deleteClosedTWServers: " + err.Error())
		return
	}
	if list == nil || len(list.Items) <= 0 {
		return
	}

	var keys []string
	for _, s := range list.Items {
		keys = append(keys, s.Key)
	}

	if len(keys) > 0 {
		deleted, err := h.observationRepo.Delete(context.Background(), &model.ObservationFilter{
			Server: keys,
		})
		if err != nil {
			log.Errorln("deleteClosedTWServers: " + err.Error())
		} else {
			log.
				WithField("numberOfDeletedObservations", len(deleted)).
				Infof("deleteClosedTWServers: some of the observations have been deleted")
		}
	}
}

func (h *handler) updateBotStatus() {
	if err := h.discord.UpdateStatus(h.status); err != nil {
		log.Error("updateBotStatus: " + err.Error())
	}
}
