

# Generate swagger file from api blueprint markdown
convert:
	@docker run --rm -v $(shell pwd):/docs cbarraford/apib2swagger -i blueprint.apib -o swagger.json

swagger: convert
	@docker run -p 80:8080 -e SWAGGER_JSON=/docs/swagger.json -v $(shell pwd):/docs swaggerapi/swagger-ui
