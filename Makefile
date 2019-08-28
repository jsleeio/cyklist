# purely for rapid rebuilding --- there is no real use case envisioned for
# non-Docker builds outside local development/iteration

cyklistctl:
	go build ./cmd/cyklistctl/
	rm -f cyklistctl
