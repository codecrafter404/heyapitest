build:
	@rm -rf out
	@mkdir out
	cd ./frontend/ && bun run build
	cd ./backend/ && go build -o ../out/test . 
