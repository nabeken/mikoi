release:
	git tag -a v$(date "+%Y%m%d%H%M%S")
	git push --tags
