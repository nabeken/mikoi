release:
	git tag -a v$(shell date "+%Y%m%d%H%M%S")
	git push --tags
