
NAME := HelloWorld

all:
	@echo "Targets: deploy describe delete"

deploy:
	gcloud functions deploy $(NAME) --runtime go111 --trigger-http

describe:
	gcloud functions describe $(NAME)

delete:
	gcloud functions delete $(NAME)
