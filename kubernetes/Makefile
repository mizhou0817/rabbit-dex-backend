# which shell to use for invoking targets
SHELL=bash

# the environment we want to operate in (build/deploy) possible choices are dev/testnet/prod
GLOBAL_PROJECT=rabbitx
GLOBAL_ENV=dev
GLOBAL_AWS_SHORT_REGION=apn1
GLOBAL_SUB_ENVIRONMENT=dev
GLOBAL_NAMING_STATIC_PART_ECR_NAME=${GLOBAL_PROJECT}-${GLOBAL_ENV}-${GLOBAL_AWS_SHORT_REGION}-${GLOBAL_SUB_ENVIRONMENT}
GLOBAL_NAMING_STATIC_PART_K8S_APP_NAME=${GLOBAL_ENV}_${GLOBAL_AWS_SHORT_REGION}_${GLOBAL_SUB_ENVIRONMENT}
# building images, configuration files and repo settings
CONFIGS_FOLDER=_configs
PLATFORM=linux/amd64
CMD=docker
IS_CICD_JOB=false
K8S_NAMESPACE=rabbitx

#Default specification that can be overwritten via CI job
CURRENT_AWS_ACCOUNT_ID=618528691313
AWS_REGION=ap-northeast-1
REPOSITORY=${CURRENT_AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
# security check to ensure that build/deploy for production cannot be by mistake.
ALREADY_CONFIRMED=0

# image tag to use when interacting (building / launching) default is latest
IMAGE_TAG=latest
GIT_SHA=$(shell git rev-parse HEAD 2>/dev/null | cut -c 1-8)

# promote golang version across CI/CD makefile and dockerfiles
GO_VERSION=1.21.5
POD=

# applications which we build/deploy by default
# this can be overridden by make APPS="tarantool grafana" build etc...
APPS= \
tarantool \
timescaledb \
centrifugo \
go-fundingservice \
go-insengine \
go-liqengine \
go-pricingservice \
go-settlementservice \
go-websocket \
go-archiver \
go-api \
grafana \
go-slipstopper \
go-referralservice \
go-dashboards \
go-profile-periodics \
#go-trading-bot \

ask_confirmation:
	@if [ "$(ALREADY_CONFIRMED)" = "1" ]; then \
		exit 0; \
	fi

	@if [ "$(GLOBAL_ENV)" = "prod" ]; then \
		echo "--> PRODUCTION ENVIRONMENT DETECTED <--"; \
		echo -n "type 'yesplease' if you want to proceed: " && read ans && [ $${ans:-N} = yesplease ]; \
		$(eval ALREADY_CONFIRMED=1) \
	fi

build: ask_confirmation
	@set -e

	@if [ "$(IS_CICD_JOB)" = "false" ] && [ "$(GLOBAL_ENV)" = "dev" ]; then \
  		make func_delete_config_folder; \
		git clone -b $(GLOBAL_SUB_ENVIRONMENT) git@gitlab.com:stripsdev/rabbit-dex-config.git --depth 1 $(CONFIGS_FOLDER); \
	elif [ "$(IS_CICD_JOB)" = "false" ] && [ "$(GLOBAL_ENV)" = "prod" ]; then \
	  	make func_delete_config_folder; \
		git clone git@gitlab.com:stripsdev/rabbit-dex-config-prod.git --depth 1 $(CONFIGS_FOLDER); \
	elif [ "$(IS_CICD_JOB)" = "true" ]; then \
	  	echo "IS_CICD_JOB value is true"; \
	else \
		echo "missing environment."; \
		exit 1; \
	fi

	@echo "$$(git log -n 1 --pretty=format:"%h / %d / $$(date -u +"%Y-%m-%dT%H:%M:%SZ")" HEAD)" > .buildinfo-rabbitx; \

	@if [ "$(GLOBAL_ENV)" = "dev" ]; then \
    	make func_build_dev; \
    elif [ "$(GLOBAL_ENV)" = "prod" ]; then \
    	make func_build_prod; \
    else \
    	echo "missing environment."; \
    	exit 1; \
    fi

push: ask_confirmation
	@set -e

	@if [ "$(GLOBAL_ENV)" = "dev" ]; then \
    	make func_push_dev; \
    elif [ "$(GLOBAL_ENV)" = "prod" ]; then \
    	make func_push_prod; \
    else \
    	echo "missing environment."; \
    	exit 1; \
    fi

down: ask_confirmation
	@set -e

	@if [ "$(GLOBAL_ENV)" = "dev" ]; then \
    	make func_down_dev; \
    elif [ "$(GLOBAL_ENV)" = "prod" ]; then \
    	make func_down_prod; \
    else \
    	echo "missing environment."; \
    	exit 1; \
    fi

up: ask_confirmation
	@set -e

	@if [ "$(GLOBAL_ENV)" = "dev" ]; then \
    	make func_up_dev; \
    elif [ "$(GLOBAL_ENV)" = "prod" ]; then \
    	make func_up_prod; \
    else \
    	echo "missing environment."; \
    	exit 1; \
    fi

get-pods:
	@PODS=$$(kubectl get pods | awk 'NR>1 {print $$1}'); \
	for pod in $$PODS; do \
		echo "$$pod $$(kubectl exec --stdin --tty $$pod -- cat /.buildinfo-rabbitx)"; \
	done

deploy: down up

ssh:
	kubectl exec -ti $$(kubectl get pods | grep $(POD) | head -n1 | awk '{print $$1}') -- bash


func_build_dev:
	@set -e

	@for app in $(APPS); do \
      	echo "-->Building the $$app application for AWS $(GLOBAL_ENV) account $(GLOBAL_SUB_ENVIRONMENT) EKS cluster... "; \
 		$(CMD) buildx build --build-arg="GO_VERSION=$(GO_VERSION)" --platform $(PLATFORM) -f $$app/Dockerfile -t $(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:latest --load ../ || exit 1; \
 		$(CMD) tag $(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:latest $(REPOSITORY)/$(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:latest || exit 1; \
 		$(CMD) tag $(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:latest $(REPOSITORY)/$(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:$(GIT_SHA) || exit 1; \
    done

func_build_prod:
	@set -e

	@for app in $(APPS); do \
      	echo "-->Building prod $$app"; \
    	$(CMD) buildx build --build-arg="GO_VERSION=$(GO_VERSION)" --platform $(PLATFORM) -f $$app/Dockerfile -t rabbitx/$$app-$(GLOBAL_ENV):latest --load ../ || exit 1; \
    	$(CMD) tag rabbitx/$$app-$(GLOBAL_ENV):latest $(REPOSITORY)/rabbitx/$$app-$(GLOBAL_ENV):latest || exit 1; \
    	$(CMD) tag rabbitx/$$app-$(GLOBAL_ENV):latest $(REPOSITORY)/rabbitx/$$app-$(GLOBAL_ENV):$(GIT_SHA) || exit 1; \
    done

func_push_dev:
	@set -e

	@for app in $(APPS); do \
  		echo "-->Pushing the $$app application for AWS $(GLOBAL_ENV) account $(GLOBAL_SUB_ENVIRONMENT) EKS cluster... "; \
		$(CMD) push $(REPOSITORY)/$(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:latest || exit 1; \
        $(CMD) push $(REPOSITORY)/$(GLOBAL_NAMING_STATIC_PART_ECR_NAME)-$$app:$(GIT_SHA) || exit 1; \
    done

func_push_prod:
	@set -e

	@for app in $(APPS); do \
  		echo "-->Pushing images prod $$app"; \
		$(CMD) push $(REPOSITORY)/rabbitx/$$app-$(GLOBAL_ENV):latest || exit 1; \
        $(CMD) push $(REPOSITORY)/rabbitx/$$app-$(GLOBAL_ENV):$(GIT_SHA) || exit 1; \
    done

func_down_dev:
	@set -e

	@for app in $(APPS); do \
		echo "-->Shutting down the $$app application on $(GLOBAL_SUB_ENVIRONMENT) EKS cluster... "; \
		kubectl kustomize $$app/overlays/$(GLOBAL_NAMING_STATIC_PART_K8S_APP_NAME) | kubectl -n $(K8S_NAMESPACE) delete --ignore-not-found=true -f -; \
	done

func_down_prod:
	@set -e

	@for app in $(APPS); do \
		echo "-->Shutting down prod $$app"; \
		kubectl kustomize $$app/overlays/$(GLOBAL_ENV) | kubectl delete --ignore-not-found=true -f - || exit 1; \
	done

func_up_dev:
	@set -e

	@for app in $(APPS); do \
    	echo "-->Bringing up the $$app application on $(GLOBAL_SUB_ENVIRONMENT) EKS cluster... "; \
		kubectl kustomize $$app/overlays/$(GLOBAL_NAMING_STATIC_PART_K8S_APP_NAME) | kubectl -n $(K8S_NAMESPACE) apply -f -; \
    done

func_up_prod:
	@set -e

	@for app in $(APPS); do \
    	echo "-->Bringing up to AWS prod account for $$app"; \
		kubectl kustomize $$app/overlays/$(GLOBAL_ENV) | kubectl apply -f -; \
    done

func_delete_config_folder:
	@set -e

	@if [ -d "$(CONFIGS_FOLDER)" ]; then \
    	rm -rf $(CONFIGS_FOLDER); \
    fi
