// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

// This test case is used to check whether BACKUP/RESTORE
// compatible with different features (eg. async-commit/one-pc)
// by checking whether we can restore the whole database no matter when we backup
// Maintainer: longfangsong <longfangsong@icloud.com>
//
// example command to run this case:
// with a tidb server running under 127.0.0.1:4000, with async-commit and one-pc on:
// ./bin/backup -tidb-server 127.0.0.1:4000 -async-commit 1 -one-pc 1
//
// This case is supposed to run forever, until an error occur or got killed
// This case should tolerant with all kinds of nemesis
// For async-commit and one-pc's calculated commit_ts related issues, which this case is originally for
// I suggest clock skew and delay on pd and tikv

package main

import (
	"context"
	"flag"
	"time"

	// use mysql
	_ "github.com/go-sql-driver/mysql"

	test_infra "github.com/pingcap/tipocket/pkg/test-infra"
	"github.com/pingcap/tipocket/pkg/verify"

	"github.com/pingcap/tipocket/cmd/util"
	"github.com/pingcap/tipocket/pkg/cluster"
	"github.com/pingcap/tipocket/pkg/control"
	"github.com/pingcap/tipocket/pkg/test-infra/fixture"
	"github.com/pingcap/tipocket/tests/backup"
)

var (
	accounts        = flag.Int("accounts", 100000, "the number of accounts")
	concurrency     = flag.Int("concurrency", 200, "concurrency worker count")
	contention      = flag.String("contention", "low", "contention level, support values: high / low, default value: low")
	backupInterval  = flag.Duration("backup-interval", 1*time.Minute, "the backup interval")
	restoreInterval = flag.Duration("restore-interval", 3*time.Minute, "the restore interval")
	dbname          = flag.String("dbname", "test", "name of database to test")
	retryLimit      = flag.Int("retry-limit", 200, "retry count")
	backupURI       = flag.String("backup-uri", "local:///tmp/backup", "where the backup file should in")

	pessimistic = flag.Bool("pessimistic", true, "use pessimistic transaction")
	replicaRead = flag.String("tidb-replica-read", "leader", "tidb_replica_read mode, support values: leader / follower / leader-and-follower, default value: leader.")
	asyncCommit = flag.Bool("async-commit", true, "whether to enable the async commit feature (default false)")
	onePC       = flag.Bool("one-pc", true, "whether to enable the one-phase commit feature (default false)")
)

func main() {
	flag.Parse()

	cfg := control.Config{
		Mode:        control.ModeSelfScheduled,
		ClientCount: 1,
		RunTime:     fixture.Context.RunTime,
		RunRound:    1,
	}

	suit := util.Suit{
		Config:   &cfg,
		Provider: cluster.NewDefaultClusterProvider(),
		ClientCreator: backup.ClientCreator{
			Cfg: backup.Config{
				NumAccounts:     *accounts,
				BackupInterval:  *backupInterval,
				RestoreInterval: *restoreInterval,
				Concurrency:     *concurrency,
				RetryLimit:      *retryLimit,
				Contention:      *contention,
				DbName:          *dbname,
				BackupURI:       *backupURI,
			},
			Features: backup.Features{
				Pessimistic: *pessimistic,
				ReplicaRead: *replicaRead,
				AsyncCommit: *asyncCommit,
				OnePC:       *onePC,
			},
		},
		NemesisGens: util.ParseNemesisGenerators(fixture.Context.Nemesis),
		VerifySuit:  verify.Suit{},
		ClusterDefs: test_infra.NewDefaultCluster(fixture.Context.Namespace, fixture.Context.Namespace,
			fixture.Context.TiDBClusterConfig),
	}
	suit.Run(context.Background())
}
