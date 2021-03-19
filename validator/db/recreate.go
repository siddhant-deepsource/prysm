package db

import (
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/validator/db/kv"
	"github.com/urfave/cli/v2"
)

// Recreate --
func Recreate(cliCtx *cli.Context) error {
	dataDir := cliCtx.String(cmd.DataDirFlag.Name)
	log.Info("Opening DB")
	validatorDB, err := kv.NewKVStore(cliCtx.Context, dataDir, &kv.Config{})
	if err != nil {
		return err
	}
	log.Info("Attempting to prune")
	return validatorDB.Recreate(cliCtx.Context)
}
