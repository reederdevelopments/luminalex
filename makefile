fmt:
	@go fmt ./...

dev:
	@tailwindcss -c ./tailwind.config.js -i ./in.css -o ./app/assets/styles/main.css
	@templ generate
	@echo "--------------------------------------------------------"
	@echo ""
	@go run ./app --dev

iframe:
	@tailwindcss -c ./tailwind.config.js -i ./in.css -o ./app/assets/styles/main.css
	@templ generate
	@echo "--------------------------------------------------------"
	@echo ""
	@go run ./iframe_server

gen:
	@tailwindcss -c ./tailwind.config.js -i ./in.css -o ./app/assets/styles/main.css --minify
	@templ generate

sha = "$(shell git rev-parse --short HEAD)"

sha:
	@echo ${sha} | pbcopy
	@echo copied '${sha}' to clipboard

projectID="df-frontend"
region="africa-south1"
serviceAccount="maoni-app@df-frontend.iam.gserviceaccount.com"
tag="${region}-docker.pkg.dev/${projectID}/maoni/maoni:${sha}"

# https://cloud.google.com/artifact-registry/docs/docker/store-docker-container-images
build-container:
	@templ generate
	@tailwindcss -c ./tailwind.config.js -i ./in.css -o ./app/assets/styles/main.css --minify
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/linux_amd64/maoni ./app
	@docker build -t ${tag} .

# If the push fails with the error below, then you must re-auth your gcloud CLI
# denied: Permission "artifactregistry.repositories.uploadArtifacts" denied on resource "projects/unifi-data-lake/locations/africa-south1/repositories/maoni" (or it may not exist)
# - gcloud auth login
# - gcloud config set project unifi-data-lake
push:
	@docker push ${tag}

auth:
	@gcloud auth login


# https://cloud.google.com/sdk/gcloud/reference/run/deploy#--set-env-vars
revise:
	@gcloud run deploy maoni \
		--image ${tag} \
		--service-account ${serviceAccount}\
		--region=${region} \
		--env-vars-file=./env.prod.yaml

deploy:
	# build-container
	# push
	# revise

# Service Account Impersonation
# https://cloud.google.com/docs/authentication/use-service-account-impersonation
# https://cloud.google.com/docs/authentication/use-service-account-impersonation#adc
creds:
	@gcloud auth application-default login --impersonate-service-account maoni-app@df-frontend.iam.gserviceaccount.com
