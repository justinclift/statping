package core

import (
	"github.com/hunterlong/statup/notifiers"
	"github.com/hunterlong/statup/source"
	"github.com/hunterlong/statup/types"
	"github.com/pkg/errors"
	"os"
	"time"
)

type PluginJSON types.PluginJSON
type PluginRepos types.PluginRepos

type Core struct {
	*types.Core
	Services []*Service
}

var (
	Configs   *types.Config
	CoreApp   *Core
	SetupMode bool
	VERSION   string
)

func init() {
	CoreApp = NewCore()
}

func NewCore() *Core {
	CoreApp = new(Core)
	CoreApp.Core = new(types.Core)
	CoreApp.Started = time.Now()
	return CoreApp
}

func InsertCore(c *Core) error {
	col := DbSession.Collection("core")
	_, err := col.Insert(c.Core)
	return err
}

func (c *Core) ToCore() *types.Core {
	return c.Core
}

func InitApp() {
	SelectCore()
	InsertNotifierDB()
	SelectAllServices()
	CheckServices()
	CoreApp.Communications = notifiers.Load()
	go DatabaseMaintence()
}

func InsertNotifierDB() error {
	if DbSession == nil {
		err := DbConnection(CoreApp.DbConnection, false, ".")
		if err != nil {
			return errors.New("database connection has not been created")
		}
	}
	notifiers.Collections = DbSession.Collection("communication")
	return nil
}

func UpdateCore(c *Core) (*Core, error) {
	res := DbSession.Collection("core").Find().Limit(1)
	err := res.Update(c.Core)
	return c, err
}

func (c Core) UsingAssets() bool {
	return source.UsingAssets
}

func (c Core) SassVars() string {
	if !source.UsingAssets {
		return ""
	}
	return source.OpenAsset(".", "scss/variables.scss")
}

func (c Core) BaseSASS() string {
	if !source.UsingAssets {
		return ""
	}
	return source.OpenAsset(".", "scss/base.scss")
}

func (c Core) MobileSASS() string {
	if !source.UsingAssets {
		return ""
	}
	return source.OpenAsset(".", "scss/mobile.scss")
}

func (c Core) AllOnline() bool {
	for _, ser := range CoreApp.Services {
		s := ser.ToService()
		if !s.Online {
			return false
		}
	}
	return true
}

func SelectLastMigration() (int64, error) {
	var c *types.Core
	if DbSession == nil {
		return 0, errors.New("Database connection has not been created yet")
	}
	err := DbSession.Collection("core").Find().One(&c)
	if err != nil {
		return 0, err
	}
	return c.MigrationId, err
}

func SelectCore() (*Core, error) {
	var c *types.Core
	exists := DbSession.Collection("core").Exists()
	if !exists {
		return nil, errors.New("core database has not been setup yet.")
	}
	err := DbSession.Collection("core").Find().One(&c)
	if err != nil {
		return nil, err
	}
	CoreApp.Core = c
	CoreApp.DbConnection = Configs.Connection
	CoreApp.Version = VERSION
	CoreApp.Services, _ = SelectAllServices()
	if os.Getenv("USE_CDN") == "true" {
		CoreApp.UseCdn = true
	}
	//store = sessions.NewCookieStore([]byte(core.ApiSecret))
	return CoreApp, err
}