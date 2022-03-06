package mysql

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/pkg/errors"
    "github.com/spf13/cast"
    driver "gorm.io/driver/mysql"
    "gorm.io/gorm"
    glogger "gorm.io/gorm/logger"
    "gorm.io/gorm/schema"

    "go-hexagonal/config"
    "go-hexagonal/util/logger"
)

/**
 * @author Rancho
 * @date 2021/12/21
 */

type IMySQL interface {
    GetDB(ctx context.Context) *gorm.DB
    SetDB(DB *gorm.DB)
    Close(ctx context.Context)
    MockClient() (*gorm.DB, sqlmock.Sqlmock)
}

type client struct {
    db *gorm.DB
}

func (c *client) GetDB(ctx context.Context) *gorm.DB {
    return c.db.WithContext(ctx)
}

func (c *client) SetDB(DB *gorm.DB) {
    c.db = DB
}

func (c *client) Close(ctx context.Context) {
    sqlDB, _ := c.GetDB(ctx).DB()
    if sqlDB != nil {
        err := sqlDB.Close()
        if err != nil {
            logger.Log.Errorf(ctx, "close mysql client fail. err: %v", errors.WithStack(err))
        }
    }
    logger.Log.Info(ctx, "mysql client closed")
}

func (c *client) MockClient() (*gorm.DB, sqlmock.Sqlmock) {
    sqlDB, mock, err := sqlmock.New()
    if err != nil {
        panic("mock MySQLClient fail, err: " + err.Error())
    }
    dialector := driver.New(driver.Config{
        Conn:       sqlDB,
        DriverName: "mysql",
    })
    // a SELECT VERSION() query will be run when gorm opens the database, so we need to expect that here
    columns := []string{"version"}
    mock.ExpectQuery("SELECT VERSION()").WithArgs().WillReturnRows(
        mock.NewRows(columns).FromCSVString("1"),
    )
    db, err := gorm.Open(dialector, &gorm.Config{})

    return db, mock
}

func NewGormDB() (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=%s",
        config.Config.MySQL.User,
        config.Config.MySQL.Password,
        config.Config.MySQL.Host,
        config.Config.MySQL.Database,
        config.Config.MySQL.CharSet,
        config.Config.MySQL.ParseTime,
        config.Config.MySQL.TimeZone,
    )

    db, err := gorm.Open(driver.Open(dsn), &gorm.Config{
        NamingStrategy: schema.NamingStrategy{
            SingularTable: true,
        },
        Logger: glogger.New(
            log.New(os.Stdout, "\r\n", log.LstdFlags),
            glogger.Config{
                SlowThreshold:             200 * time.Millisecond,
                LogLevel:                  glogger.Info,
                IgnoreRecordNotFoundError: false,
                Colorful:                  true,
            }),
    })
    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    sqlDB.SetMaxIdleConns(config.Config.MySQL.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.Config.MySQL.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(cast.ToDuration(config.Config.MySQL.MaxLifeTime))
    sqlDB.SetConnMaxIdleTime(cast.ToDuration(config.Config.MySQL.MaxIdleTime))

    return db, nil
}

func NewMySQLClient() IMySQL {
    db, err := NewGormDB()
    if err != nil {
        panic(err)
    }
    return &client{db: db}
}
