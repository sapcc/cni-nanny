################################################################################
# This file is AUTOGENERATED with <https://github.com/sapcc/go-makefile-maker> #
# Edit Makefile.maker.yaml instead.                                            #
################################################################################

# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company
# SPDX-License-Identifier: Apache-2.0

MAKEFLAGS=--warn-undefined-variables
# /bin/sh is dash on Debian which does not support all features of ash/bash
# to fix that we use /bin/bash only on Debian to not break Alpine
ifneq (,$(wildcard /etc/os-release)) # check file existence
	ifneq ($(shell grep -c debian /etc/os-release),0)
		SHELL := /bin/bash
	endif
endif
UNAME_S := $(shell uname -s)
SED = sed
XARGS = xargs
ifeq ($(UNAME_S),Darwin)
	SED = gsed
	XARGS = gxargs
endif

default: FORCE
	@echo 'There is nothing to build, use `make check` for running the test suite or `make help` for a list of available targets.'

install-goimports: FORCE
	@if ! hash goimports 2>/dev/null; then printf "\e[1;36m>> Installing goimports (this may take a while)...\e[0m\n"; go install golang.org/x/tools/cmd/goimports@latest; fi

install-golangci-lint: FORCE
	@if ! hash golangci-lint 2>/dev/null; then printf "\e[1;36m>> Installing golangci-lint (this may take a while)...\e[0m\n"; go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; fi

install-modernize: FORCE
	@if ! hash modernize 2>/dev/null; then printf "\e[1;36m>> Installing modernize (this may take a while)...\e[0m\n"; go install golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest; fi

install-shellcheck: FORCE
	@if ! hash shellcheck 2>/dev/null; then printf "\e[1;36m>> Installing shellcheck...\e[0m\n"; SHELLCHECK_ARCH=$(shell uname -m); SHELLCHECK_OS=$(shell uname -s | tr '[:upper:]' '[:lower:]'); if [[ "$$SHELLCHECK_OS" == "darwin" ]]; then SHELLCHECK_OS=macos; fi; SHELLCHECK_VERSION="stable"; if command -v curl >/dev/null 2>&1; then GET="curl -sLo-"; elif command -v wget >/dev/null 2>&1; then GET="wget -O-"; else echo "Didn't find curl or wget to download shellcheck"; exit 2; fi; $$GET "https://github.com/koalaman/shellcheck/releases/download/$$SHELLCHECK_VERSION/shellcheck-$$SHELLCHECK_VERSION.$$SHELLCHECK_OS.$$SHELLCHECK_ARCH.tar.xz" | tar -Jxf -; BIN=$$(go env GOBIN); if [[ -z $$BIN ]]; then BIN=$$(go env GOPATH)/bin; fi; install -Dm755 shellcheck-$$SHELLCHECK_VERSION/shellcheck -t "$$BIN"; rm -rf shellcheck-$$SHELLCHECK_VERSION; fi

install-go-licence-detector: FORCE
	@if ! hash go-licence-detector 2>/dev/null; then printf "\e[1;36m>> Installing go-licence-detector (this may take a while)...\e[0m\n"; go install go.elastic.co/go-licence-detector@latest; fi

install-addlicense: FORCE
	@if ! hash addlicense 2>/dev/null; then printf "\e[1;36m>> Installing addlicense (this may take a while)...\e[0m\n"; go install github.com/google/addlicense@latest; fi

install-reuse: FORCE
	@if ! hash reuse 2>/dev/null; then if ! hash pip3 2>/dev/null; then printf "\e[1;31m>> Cannot install reuse because no pip3 was found. Either install it using your package manager or install pip3\e[0m\n"; else printf "\e[1;36m>> Installing reuse...\e[0m\n"; pip3 install --user reuse; fi; fi

prepare-static-check: FORCE install-golangci-lint install-modernize install-shellcheck install-go-licence-detector install-addlicense install-reuse

install-controller-gen: FORCE
	@if ! hash controller-gen 2>/dev/null; then printf "\e[1;36m>> Installing controller-gen (this may take a while)...\e[0m\n"; go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest; fi

install-setup-envtest: FORCE
	@if ! hash setup-envtest 2>/dev/null; then printf "\e[1;36m>> Installing setup-envtest (this may take a while)...\e[0m\n"; go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest; fi

GO_BUILDFLAGS =
GO_LDFLAGS =
GO_TESTENV =
GO_BUILDENV =

# which packages to test with test runner
GO_TESTPKGS := $(shell go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.Dir}}{{end}}' ./...)
ifeq ($(GO_TESTPKGS),)
GO_TESTPKGS := ./...
endif
# which packages to measure coverage for
GO_COVERPKGS := $(shell go list ./...)
# to get around weird Makefile syntax restrictions, we need variables containing nothing, a space and comma
null :=
space := $(null) $(null)
comma := ,

check: FORCE static-check build/cover.html
	@printf "\e[1;32m>> All checks successful.\e[0m\n"

generate: install-controller-gen
	@printf "\e[1;36m>> controller-gen\e[0m\n"
	@controller-gen crd rbac:roleName=cni-nanny webhook paths="./..." output:crd:artifacts:config=crd
	@controller-gen object paths="./..."

run-golangci-lint: FORCE install-golangci-lint
	@printf "\e[1;36m>> golangci-lint\e[0m\n"
	@golangci-lint config verify
	@golangci-lint run

run-modernize: FORCE install-modernize
	@printf "\e[1;36m>> modernize\e[0m\n"
	@modernize $(GO_TESTPKGS)

run-shellcheck: FORCE install-shellcheck
	@printf "\e[1;36m>> shellcheck\e[0m\n"
	@find .  -type f \( -name '*.bash' -o -name '*.ksh' -o -name '*.zsh' -o -name '*.sh' -o -name '*.shlib' \) -exec shellcheck  {} +

build/cover.out: FORCE generate install-setup-envtest | build
	@printf "\e[1;36m>> Running tests\e[0m\n"
	KUBEBUILDER_ASSETS=$$(setup-envtest use 1.32 -p path) go run github.com/onsi/ginkgo/v2/ginkgo run --randomize-all -output-dir=build $(GO_BUILDFLAGS) -ldflags '-s -w $(GO_LDFLAGS)' -covermode=count -coverpkg=$(subst $(space),$(comma),$(GO_COVERPKGS)) $(GO_TESTPKGS)
	@mv build/coverprofile.out build/cover.out

build/cover.html: build/cover.out
	@printf "\e[1;36m>> go tool cover > build/cover.html\e[0m\n"
	@go tool cover -html $< -o $@

check-addlicense: FORCE install-addlicense
	@printf "\e[1;36m>> addlicense --check\e[0m\n"
	@addlicense --check -- $(patsubst $(shell awk '$$1 == "module" {print $$2}' go.mod)%,.%/*.go,$(shell go list ./...))

check-reuse: FORCE install-reuse
	@printf "\e[1;36m>> reuse lint\e[0m\n"
	@if ! reuse lint -q; then reuse lint; fi

check-license-headers: FORCE check-addlicense check-reuse

__static-check: FORCE run-shellcheck run-golangci-lint run-modernize check-dependency-licenses check-license-headers

static-check: FORCE
	@$(MAKE) --keep-going --no-print-directory __static-check

build:
	@mkdir $@

tidy-deps: FORCE
	go mod tidy
	go mod verify

license-headers: FORCE install-addlicense install-reuse
	@printf "\e[1;36m>> addlicense (for license headers on source code files)\e[0m\n"
	@printf "%s\0" $(patsubst $(shell awk '$$1 == "module" {print $$2}' go.mod)%,.%/*.go,$(shell go list ./...)) | $(XARGS) -0 -I{} bash -c 'year="$$(grep 'Copyright' {} | head -n1 | grep -E -o '"'"'[0-9]{4}(-[0-9]{4})?'"'"')"; if [[ -z "$$year" ]]; then year=$$(date +%Y); fi; gawk -i inplace '"'"'{if (display) {print} else {!/^\/\*/ && !/^\*/}}; {if (!display && $$0 ~ /^(package |$$)/) {display=1} else { }}'"'"' {}; addlicense -c "SAP SE or an SAP affiliate company" -s=only -y "$$year" -- {}; $(SED) -i '"'"'1s+// Copyright +// SPDX-FileCopyrightText: +'"'"' {}; '
	@printf "\e[1;36m>> reuse annotate (for license headers on other files)\e[0m\n"
	@reuse lint -j | jq -r '.non_compliant.missing_licensing_info[]' | grep -vw vendor | $(XARGS) reuse annotate -c 'SAP SE or an SAP affiliate company' -l Apache-2.0 --skip-unrecognised
	@printf "\e[1;36m>> reuse download --all\e[0m\n"
	@reuse download --all
	@printf "\e[1;35mPlease review the changes. If *.license files were generated, consider instructing go-makefile-maker to add overrides to REUSE.toml instead.\e[0m\n"

check-dependency-licenses: FORCE install-go-licence-detector
	@printf "\e[1;36m>> go-licence-detector\e[0m\n"
	@go list -m -mod=readonly -json all | go-licence-detector -includeIndirect -rules .license-scan-rules.json -overrides .license-scan-overrides.jsonl

goimports: FORCE install-goimports
	@printf "\e[1;36m>> goimports -w -local https://github.com/sapcc/cni-nanny\e[0m\n"
	@goimports -w -local github.com/sapcc/cni-nanny $(patsubst $(shell awk '$$1 == "module" {print $$2}' go.mod)%,.%/*.go,$(shell go list ./...))

modernize: FORCE install-modernize
	@printf "\e[1;36m>> modernize -fix ./...\e[0m\n"
	@modernize -fix ./...

clean: FORCE
	git clean -dxf build

vars: FORCE
	@printf "GO_BUILDFLAGS=$(GO_BUILDFLAGS)\n"
	@printf "GO_COVERPKGS=$(GO_COVERPKGS)\n"
	@printf "GO_LDFLAGS=$(GO_LDFLAGS)\n"
	@printf "GO_TESTPKGS=$(GO_TESTPKGS)\n"
	@printf "MAKE=$(MAKE)\n"
	@printf "SED=$(SED)\n"
	@printf "UNAME_S=$(UNAME_S)\n"
	@printf "XARGS=$(XARGS)\n"
help: FORCE
	@printf "\n"
	@printf "\e[1mUsage:\e[0m\n"
	@printf "  make \e[36m<target>\e[0m\n"
	@printf "\n"
	@printf "\e[1mGeneral\e[0m\n"
	@printf "  \e[36mvars\e[0m                         Display values of relevant Makefile variables.\n"
	@printf "  \e[36mhelp\e[0m                         Display this help.\n"
	@printf "\n"
	@printf "\e[1mPrepare\e[0m\n"
	@printf "  \e[36minstall-goimports\e[0m            Install goimports required by goimports/static-check\n"
	@printf "  \e[36minstall-golangci-lint\e[0m        Install golangci-lint required by run-golangci-lint/static-check\n"
	@printf "  \e[36minstall-modernize\e[0m            Install modernize required by run-modernize/static-check\n"
	@printf "  \e[36minstall-shellcheck\e[0m           Install shellcheck required by run-shellcheck/static-check\n"
	@printf "  \e[36minstall-go-licence-detector\e[0m  Install-go-licence-detector required by check-dependency-licenses/static-check\n"
	@printf "  \e[36minstall-addlicense\e[0m           Install addlicense required by check-license-headers/license-headers/static-check\n"
	@printf "  \e[36minstall-reuse\e[0m                Install reuse required by license-headers/check-reuse\n"
	@printf "  \e[36mprepare-static-check\e[0m         Install any tools required by static-check. This is used in CI before dropping privileges, you should probably install all the tools using your package manager\n"
	@printf "  \e[36minstall-controller-gen\e[0m       Install controller-gen required by static-check and build-all. This is used in CI before dropping privileges, you should probably install all the tools using your package manager\n"
	@printf "  \e[36minstall-setup-envtest\e[0m        Install setup-envtest required by check. This is used in CI before dropping privileges, you should probably install all the tools using your package manager\n"
	@printf "\n"
	@printf "\e[1mTest\e[0m\n"
	@printf "  \e[36mcheck\e[0m                        Run the test suite (unit tests and golangci-lint).\n"
	@printf "  \e[36mgenerate\e[0m                     Generate code for Kubernetes CRDs and deepcopy.\n"
	@printf "  \e[36mrun-golangci-lint\e[0m            Install and run golangci-lint. Installing is used in CI, but you should probably install golangci-lint using your package manager.\n"
	@printf "  \e[36mrun-modernize\e[0m                Install and run modernize. Installing is used in CI, but you should probably install modernize using your package manager.\n"
	@printf "  \e[36mrun-shellcheck\e[0m               Install and run shellcheck. Installing is used in CI, but you should probably install shellcheck using your package manager.\n"
	@printf "  \e[36mbuild/cover.out\e[0m              Run tests and generate coverage report.\n"
	@printf "  \e[36mbuild/cover.html\e[0m             Generate an HTML file with source code annotations from the coverage report.\n"
	@printf "  \e[36mcheck-addlicense\e[0m             Check license headers in all non-vendored .go files with addlicense.\n"
	@printf "  \e[36mcheck-reuse\e[0m                  Check reuse compliance\n"
	@printf "  \e[36mcheck-license-headers\e[0m        Run static code checks\n"
	@printf "  \e[36mstatic-check\e[0m                 Run static code checks\n"
	@printf "\n"
	@printf "\e[1mDevelopment\e[0m\n"
	@printf "  \e[36mtidy-deps\e[0m                    Run go mod tidy and go mod verify.\n"
	@printf "  \e[36mlicense-headers\e[0m              Add (or overwrite) license headers on all non-vendored source code files.\n"
	@printf "  \e[36mcheck-dependency-licenses\e[0m    Check all dependency licenses using go-licence-detector.\n"
	@printf "  \e[36mgoimports\e[0m                    Run goimports on all non-vendored .go files\n"
	@printf "  \e[36mmodernize\e[0m                    Run modernize on all non-vendored .go files\n"
	@printf "  \e[36mclean\e[0m                        Run git clean.\n"

.PHONY: FORCE
