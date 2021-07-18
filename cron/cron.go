package cron

import (
	"context"
	"fmt"
	"github.com/Kichiyaki/appmode"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/model"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/util/twutil"
)

const (
	colorLostVillage      = 0xff0000
	colorConqueredVillage = 0x00ff00
)

var log = logrus.WithField("package", "cron")

type Config struct {
	ServerRepo      server.Repository
	ObservationRepo observation.Repository
	Discord         *discord.Session
	GroupRepo       group.Repository
	API             *sdk.SDK
	Status          string
}

type Cron struct {
	*cron.Cron
	lastEnnoblementAt map[string]time.Time
	serverRepo        server.Repository
	observationRepo   observation.Repository
	groupRepo         group.Repository
	discord           *discord.Session
	api               *sdk.SDK
	status            string
}

func New(cfg Config) *Cron {
	c := &Cron{
		Cron: cron.New(
			cron.WithChain(
				cron.SkipIfStillRunning(
					cron.PrintfLogger(log),
				),
			),
		),
		lastEnnoblementAt: make(map[string]time.Time),
		serverRepo:        cfg.ServerRepo,
		observationRepo:   cfg.ObservationRepo,
		groupRepo:         cfg.GroupRepo,
		discord:           cfg.Discord,
		api:               cfg.API,
		status:            cfg.Status,
	}

	checkEnnoblements := trackDuration(log, c.checkEnnoblements, "checkEnnoblements")
	checkBotServers := trackDuration(log, c.checkBotServers, "checkBotServers")
	deleteClosedTribalWarsServers := trackDuration(log,
		c.deleteClosedTWServers,
		"deleteClosedTWServers")
	updateBotStatus := trackDuration(log, c.updateBotStatus, "updateBotStatus")
	c.AddFunc("@every 1m", checkEnnoblements)
	c.AddFunc("@every 30m", checkBotServers)
	c.AddFunc("@every 2h10m", deleteClosedTribalWarsServers)
	c.AddFunc("@every 2h", updateBotStatus)
	go func() {
		checkBotServers()
		deleteClosedTribalWarsServers()
		updateBotStatus()
		if appmode.Equals(appmode.DevelopmentMode) {
			checkEnnoblements()
		}
	}()

	return c
}

func (c *Cron) loadEnnoblements(servers []string) (map[string]ennoblements, error) {
	m := make(map[string]ennoblements)

	if len(servers) == 0 {
		return m, nil
	}

	query := ""

	for _, s := range servers {
		lastEnnoblementAt, ok := c.lastEnnoblementAt[s]
		if !ok {
			lastEnnoblementAt = time.Now().Add(-1 * time.Minute)
			c.lastEnnoblementAt[s] = lastEnnoblementAt
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
	if err := c.api.Post(fmt.Sprintf(`query { %s }`, query), &resp); err != nil {
		return m, errors.Wrap(err, "loadEnnoblements")
	}

	for s, singleServerResp := range resp {
		if singleServerResp == nil {
			continue
		}
		m[s] = singleServerResp.Items
		lastEnnoblement := m[s].getLastEnnoblement()
		if lastEnnoblement != nil {
			c.lastEnnoblementAt[s] = lastEnnoblement.EnnobledAt
		}
	}

	return m, nil
}

func (c *Cron) loadVersions(servers []string) ([]*twmodel.Version, error) {
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

	versionList, err := c.api.Version.Browse(0, 0, []string{"code ASC"}, &twmodel.VersionFilter{
		Code: versionCodes,
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load versions")
	}

	return versionList.Items, nil
}

func (c *Cron) checkEnnoblements() {
	servers, err := c.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Errorln("checkEnnoblements:", err.Error())
		return
	}
	log.
		WithField("servers", servers).
		Info("checkEnnoblements: servers have been loaded")

	groups, total, err := c.groupRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Errorln("checkEnnoblements:", err.Error())
		return
	}
	log.
		WithField("numberOfGroups", total).
		Info("checkEnnoblements: groups have been loaded")

	versions, err := c.loadVersions(servers)
	if err != nil {
		log.Errorln("checkEnnoblements:", err)
		return
	}
	log.
		WithField("numberOfVersions", len(versions)).
		Info("checkEnnoblements: versions have been loaded")

	ennoblementsByServerKey, err := c.loadEnnoblements(servers)
	if err != nil {
		log.Errorln("checkEnnoblements:", err)
	}
	log.Info("checkEnnoblements: ennoblements have been loaded")

	for _, g := range groups {
		if g.ConqueredVillagesChannelID == "" && g.LostVillagesChannelID == "" {
			continue
		}
		localizer := message.NewLocalizer(g.Server.Lang)
		lostVillagesBldr := &discord.MessageEmbedFieldBuilder{}
		conqueredVillagesBldr := &discord.MessageEmbedFieldBuilder{}
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
						newMsgDataConfig := newEnnoblementMsgConfig{
							host:        version.Host,
							server:      obs.Server,
							ennoblement: ennoblement,
							localizer:   localizer,
						}
						lostVillagesBldr.Append(newEnnoblementMsg(newMsgDataConfig).String())
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

						newMsgDataConfig := newEnnoblementMsgConfig{
							host:        version.Host,
							server:      obs.Server,
							ennoblement: ennoblement,
							localizer:   localizer,
						}
						conqueredVillagesBldr.Append(newEnnoblementMsg(newMsgDataConfig).String())
					}
				}
			}
		}

		timestamp := time.Now().Format(time.RFC3339)
		if !conqueredVillagesBldr.IsEmpty() {
			title := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CronConqueredVillagesTitle,
			})
			conqueredVillagesBldr.SetName(title)
			go c.discord.SendEmbed(g.ConqueredVillagesChannelID,
				discord.
					NewEmbed().
					SetTitle(title).
					SetColor(colorConqueredVillage).
					SetFields(conqueredVillagesBldr.ToMessageEmbedFields()).
					SetTimestamp(timestamp))
		}

		if !lostVillagesBldr.IsEmpty() {
			title := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CronLostVillagesTitle,
			})
			lostVillagesBldr.SetName(title)
			go c.discord.SendEmbed(g.LostVillagesChannelID,
				discord.
					NewEmbed().
					SetTitle(title).
					SetColor(colorLostVillage).
					SetFields(lostVillagesBldr.ToMessageEmbedFields()).
					SetTimestamp(timestamp))
		}
	}
}

func (c *Cron) checkBotServers() {
	servers, total, err := c.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Error("checkBotServers: " + err.Error())
		return
	}
	log.
		WithField("numberOfServers", total).
		Info("checkBotServers: loaded servers")

	var idsToDelete []string
	for _, s := range servers {
		if isGuildMember, _ := c.discord.IsGuildMember(s.ID); !isGuildMember {
			idsToDelete = append(idsToDelete, s.ID)
		}
	}

	if len(idsToDelete) > 0 {
		deleted, err := c.serverRepo.Delete(context.Background(), &model.ServerFilter{
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

func (c *Cron) deleteClosedTWServers() {
	servers, err := c.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Error("deleteClosedTWServers: " + err.Error())
		return
	}
	log.
		WithField("servers", servers).
		Info("deleteClosedTWServers: loaded servers")

	list, err := c.api.Server.Browse(0, 0, []string{"key ASC"}, &twmodel.ServerFilter{
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
		deleted, err := c.observationRepo.Delete(context.Background(), &model.ObservationFilter{
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

func (c *Cron) updateBotStatus() {
	if err := c.discord.UpdateStatus(c.status); err != nil {
		log.Error("updateBotStatus: " + err.Error())
	}
}
