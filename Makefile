.PHONY: source_file

source_file:
	mv game/main.go game/main.go.orig
	cat game/*.go > game/main.go
