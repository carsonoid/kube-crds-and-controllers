PHONY: deps clean gen-go-crds

deps:
	glide i

clean: clean-build clean-go-crds

clean-build:
	# Clean up build dir
	rm -rf build/*

clean-go-crds:
	# Delete all generated code.
	rm -rf controllers/crd-configured/pkg/client
	rm -f  controllers/crd-configured/pkg/apis/*/*/zz_generated.deepcopy.go

gen-go-crds: clean-go-crds # Must be cleaned so that removals happen properly
	./vendor/k8s.io/code-generator/generate-groups.sh \
	all \
	github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/client \
	github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/apis \
	"podlabeler:v1alpha1"

	# workaround https://github.com/openshift/origin/issues/10357
	find controllers/*/pkg/client -name "clientset_generated.go" -exec sed -i'' 's/return \\&Clientset{fakePtr/return \\&Clientset{\\&fakePtr/g' '{}' \;

controllers/%:
	@mkdir build >/dev/null 2>&1|| true
	go build -i -o build/$@ $@.go

run-controllers/%: controllers/%
	./build/controllers/$* $OPTS