PHONY: all gen-go-crds all-controllers

gen-go-crds: clean-go-crds # Must be cleaned so that removals happen properly
	./vendor/k8s.io/code-generator/generate-groups.sh \
	all \
	github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/client \
	github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/apis \
	"podlabeler:v1alpha1"

	# workaround https://github.com/openshift/origin/issues/10357
	find controllers/*/pkg/client -name "clientset_generated.go" -exec sed -i'' 's/return \\&Clientset{fakePtr/return \\&Clientset{\\&fakePtr/g' '{}' \;

all: clean godeps gen-go-crds all-controllers

all-controllers: \
controllers/hard-coded/simple \
controllers/hard-coded/structured \
controllers/configmap-configured/single-config \
controllers/configmap-configured/multi-config \
controllers/crd-configured/simple \
controllers/crd-configured/workqueue

help:
	@echo "make run-controllers/hard-coded/simple"
	@echo "make run-controllers/hard-coded/structured"
	@echo "make run-controllers/configmap-configured/single-config"
	@echo "make run-controllers/configmap-configured/multi-config"
	@echo "make run-controllers/crd-configured/simple"
	@echo "make run-controllers/crd-configured/workqueue"

controllers/%:
	@mkdir build >/dev/null 2>1|| true
	go build -i -o build/$@ $@.go

run-controllers/%: controllers/%
	./build/controllers/$*

godeps:
	glide i

clean: clean-build clean-go-crds

clean-build:
	# Clean up build idr
	rm -rf build/*

clean-go-crds:
	# Delete all generated code.
	rm -rf controllers/crd-configured/pkg/client
	rm -f  controllers/crd-configured/pkg/apis/*/*/zz_generated.deepcopy.go
