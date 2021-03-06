/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// A set of common functions needed by cmd/kubectl and pkg/kubectl packages.
package kubectl

import (
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const kubectlAnnotationPrefix = "kubectl.kubernetes.io/"

type NamespaceInfo struct {
	Namespace string
}

func listOfImages(spec *api.PodSpec) []string {
	var images []string
	for _, container := range spec.Containers {
		images = append(images, container.Image)
	}
	return images
}

func makeImageList(spec *api.PodSpec) string {
	return strings.Join(listOfImages(spec), ",")
}

// OutputVersionMapper is a RESTMapper that will prefer mappings that
// correspond to a preferred output version (if feasible)
type OutputVersionMapper struct {
	meta.RESTMapper

	// output versions takes a list of preferred GroupVersions. Only the first
	// hit for a given group will have effect.  This allows different output versions
	// depending upon the group of the kind being requested
	OutputVersions []unversioned.GroupVersion
}

// RESTMapping implements meta.RESTMapper by prepending the output version to the preferred version list.
func (m OutputVersionMapper) RESTMapping(gk unversioned.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	for _, preferredVersion := range m.OutputVersions {
		if gk.Group == preferredVersion.Group {
			mapping, err := m.RESTMapper.RESTMapping(gk, preferredVersion.Version)
			if err == nil {
				return mapping, nil
			}

			break
		}
	}

	return m.RESTMapper.RESTMapping(gk, versions...)
}

// ShortcutExpander is a RESTMapper that can be used for Kubernetes
// resources.
type ShortcutExpander struct {
	meta.RESTMapper
}

var _ meta.RESTMapper = &ShortcutExpander{}

// KindFor implements meta.RESTMapper. It expands the resource first, then invokes the wrapped
// mapper.
func (e ShortcutExpander) KindFor(resource string) (unversioned.GroupVersionKind, error) {
	resource = expandResourceShortcut(resource)
	return e.RESTMapper.KindFor(resource)
}

// ResourceIsValid takes a string (kind) and checks if it's a valid resource.
// It expands the resource first, then invokes the wrapped mapper.
func (e ShortcutExpander) ResourceIsValid(resource string) bool {
	return e.RESTMapper.ResourceIsValid(expandResourceShortcut(resource))
}

// ResourceSingularizer expands the named resource and then singularizes it.
func (e ShortcutExpander) ResourceSingularizer(resource string) (string, error) {
	return e.RESTMapper.ResourceSingularizer(expandResourceShortcut(resource))
}

// expandResourceShortcut will return the expanded version of resource
// (something that a pkg/api/meta.RESTMapper can understand), if it is
// indeed a shortcut. Otherwise, will return resource unmodified.
func expandResourceShortcut(resource string) string {
	shortForms := map[string]string{
		// Please keep this alphabetized
		"cs":     "componentstatuses",
		"ds":     "daemonsets",
		"ep":     "endpoints",
		"ev":     "events",
		"hpa":    "horizontalpodautoscalers",
		"ing":    "ingresses",
		"limits": "limitranges",
		"no":     "nodes",
		"ns":     "namespaces",
		"po":     "pods",
		"pvc":    "persistentvolumeclaims",
		"pv":     "persistentvolumes",
		"quota":  "resourcequotas",
		"rc":     "replicationcontrollers",
		"svc":    "services",
	}
	if expanded, ok := shortForms[resource]; ok {
		return expanded
	}
	return resource
}
