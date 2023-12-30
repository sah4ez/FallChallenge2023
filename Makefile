.PHONY: source_file

source_file:
	mv game/main.go game/main.go.orig
	cat src/*.go | grep package | uniq > game/main.go
	@echo "import (" >> game/main.go
	cat src/*.go | awk '/import \(/,/\)/' | sed '/import (/d;/)/d' | sort | uniq >> game/main.go
	@echo ")" >> game/main.go
	cat src/*.go | egrep -v "(package|import|^\)|	\")" >> game/main.go
	goimports -w game/main.go    
