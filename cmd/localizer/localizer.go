// Copyright 2020 Jared Allard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"os"

	"github.com/jaredallard/localizer/internal/kube"
	"github.com/jaredallard/localizer/internal/proxier"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()
	log := logrus.New().WithContext(ctx)

	app := cli.App{
		Version: "1.0.0",
		Name:    "localizer",
		Flags: []cli.Flag{
			// Note: KUBECONFIG is respected, but we don't allow passing a
			// CLI argument to reduce the complexity and re-parsing of it.
			&cli.StringFlag{
				Name:    "context",
				Usage:   "Specify Kubernetes context to use",
				EnvVars: []string{"KUBECONTEXT"},
			},
		},
		Action: func(c *cli.Context) error {
			_, k, err := kube.GetKubeClient(c.String("context"))
			if err != nil {
				return errors.Wrap(err, "failed to create kube client")
			}

			d := proxier.NewDiscoverer(k, log)
			p := proxier.NewProxier(k, log)

			services, err := d.Discover(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to discover services")
			}

			if len(services) == 0 {
				log.Info("found no services, exiting ...")
				return nil
			}

			if err := p.Add(services...); err != nil {
				return errors.Wrap(err, "failed to add discovered services to proxy")
			}

			return errors.Wrap(p.Proxy(ctx), "failed to start proxier")
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Errorf("failed to run: %v", err)
		os.Exit(1)
	}
}