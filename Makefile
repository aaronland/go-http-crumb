local-scan:
	/usr/local/bin/sonar-scanner/bin/sonar-scanner -Dsonar.projectKey=go-http-crumb -Dsonar.sources=. -Dsonar.host.url=http://localhost:9000 -Dsonar.login=$(TOKEN)
