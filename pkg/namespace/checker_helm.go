package namespace

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type HelmChecker struct {
	actionConfig *action.Configuration
}

const (
	driverSecret    = "secret"
	helmCheckerName = "Helm"
)

func NewHelmChecker() *HelmChecker {
	return &HelmChecker{actionConfig: new(action.Configuration)}
}

func (h *HelmChecker) Name() string {
	return helmCheckerName
}

func (h *HelmChecker) NonEmptyNamespaces(_ context.Context, nonEmptyNamespaces map[string]struct{}) error {
	if err := h.actionConfig.Init(genericclioptions.NewConfigFlags(true), "", driverSecret, func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format, v))
	}); err != nil {
		return err
	}

	listAction := action.NewList(h.actionConfig)
	listAction.AllNamespaces = true
	releases, err := listAction.Run()
	if err != nil {
		return err
	}

	for _, release := range releases {
		if release.Info.Deleted.IsZero() {
			// Found an active Release in this namespace
			nonEmptyNamespaces[release.Namespace] = struct{}{}
		}
	}
	return nil
}
