# Aldriguz AIO

A Go framework originally built to add site-specific automation modules. The main implementation in this repository is a Queue-it handler; the other menu entries and the SkateDeluxe package are incomplete examples of how additional sites could be organized.

## Queue-it handler

The Queue-it module reads task URLs from CSV, uses a configured proxy, extracts queue and challenge metadata, and follows the waiting-room flow. It includes handlers for:

- proof-of-work challenges;
- invisible reCAPTCHA through a user-supplied 2Captcha API key;
- Queue-it BotDetect image challenges, including image classification and OCR;
- enqueue requests, queue-status polling, and redirect URL retrieval.

It can send a completed queue link to a user-supplied Discord success webhook. Challenge formats and target integrations may change; this repository is an archived prototype rather than a maintained or production-ready service. Use it only where you are authorized to automate access.

## Configuration

Requires Go 1.20 or newer. The image CAPTCHA code uses `gosseract`, so building that package also requires the native Tesseract and Leptonica development libraries.

1. Copy `settings/settings.example.json` to `settings/settings.json` and fill in your own license, 2Captcha API key, and optional Discord webhook URLs.
2. Set `WHOP_API_TOKEN` in your local environment. The existing startup flow uses this token to validate the license through Whop. The optional process-alert component uses `ALDRIGUZ_ALERT_WEBHOOK` if enabled.
3. Copy `tasks/queue.example.csv` to a new CSV file under `tasks/` and replace the example URL with a queue URL you are permitted to use. The third column is the task delay in seconds.
4. Add a local proxy file under `proxies/`, with one `host:port:username:password` entry per line.
5. Run `go run .` and choose **Queue IT** from the menu.

The program creates its runtime folders if they are missing. The local settings, task files, proxies, accounts, data, orders, and environment files are ignored by Git. The files ending in `.example.json` and `.example.csv` are placeholders and contain no credentials.

## Project layout

- `main.go` and `packages/modules/`: startup and module selection.
- `modules/queue/`: Queue-it flow, challenge handlers, and status polling.
- `modules/skatedeluxe/`: unfinished example of a site-specific module.
- `packages/auth/`: Whop license validation using an environment-provided token.
- `packages/captchaSolver/`: 2Captcha integration.
- `packages/reader/`, `packages/proxy/`, and `packages/discord/`: local configuration, proxies, and webhook delivery.

To add another site, create a module under `modules/`, register it in the selector, and connect its task loader in `main.go`. Do not commit real credentials or task data. If you publish a repository that previously contained secrets, use a new history and revoke those credentials before publishing.
