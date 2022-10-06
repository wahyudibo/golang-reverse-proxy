#!/usr/bin/env sh

docker run -d -p 9222:9222 --rm --name headless-shell --init chromedp/headless-shell:latest chromedp/headless-shell --user-agent="${PROXY_USER_AGENT}" --no-sandbox --disable-gpu