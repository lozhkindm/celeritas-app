package celeritas

import (
	"path"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop"
)

func (c *Celeritas) PopConnect() (*pop.Connection, error) {
	tx, err := pop.Connect("development")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *Celeritas) CreatePopMigration(up, down []byte, name, ext string) error {
	if err := pop.MigrationCreate(path.Join(c.RootPath, "migrations"), name, ext, up, down); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) MigrateUp(tx *pop.Connection) error {
	migrator, err := pop.NewFileMigrator(path.Join(c.RootPath, "migrations"), tx)
	if err != nil {
		return err
	}
	if err := migrator.Up(); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) MigrateDown(tx *pop.Connection, steps ...int) error {
	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}
	migrator, err := pop.NewFileMigrator(path.Join(c.RootPath, "migrations"), tx)
	if err != nil {
		return err
	}
	if err := migrator.Down(step); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) MigrateReset(tx *pop.Connection) error {
	migrator, err := pop.NewFileMigrator(path.Join(c.RootPath, "migrations"), tx)
	if err != nil {
		return err
	}
	if err := migrator.Reset(); err != nil {
		return err
	}
	return nil
}
