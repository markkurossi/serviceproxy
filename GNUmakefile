
NAME := ServiceProxy

all:
	@echo "Targets: deploy describe delete"

# Now the entry point is the function name $(NAME). But you can alse
# select another function with the --entry-point=NAME option.

deploy:
	gcloud functions deploy $(NAME) --runtime go113 --trigger-http

describe:
	gcloud functions describe $(NAME)

delete:
	gcloud functions delete $(NAME)
