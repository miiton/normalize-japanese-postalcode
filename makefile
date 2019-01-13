build:
	GOOS=windows GOARCH=386 go build -o njp.exe && \
	7z a normalize-japanese-postalcode_win32.zip njp.exe && \
	rm njp.exe
	GOOS=windows GOARCH=amd64 go build -o njp.exe && \
	7z a normalize-japanese-postalcode_win64.zip njp.exe && \
	rm njp.exe
	GOOS=linux GOARCH=386 go build -o njp && \
	7z a normalize-japanese-postalcode_linux32.zip njp && \
	rm njp
	GOOS=linux GOARCH=amd64 go build -o njp && \
	7z a normalize-japanese-postalcode_linux64.zip njp && \
	rm njp
	GOOS=darwin GOARCH=386 go build -o njp && \
	7z a normalize-japanese-postalcode_mac32.zip njp && \
	rm njp
	GOOS=darwin GOARCH=amd64 go build -o njp && \
	7z a normalize-japanese-postalcode_mac64.zip njp && \
	rm njp
