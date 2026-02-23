---
slug: deployment
category: operations
generatedAt: 2026-02-23T15:50:10.895Z
relevantFiles:
  - docker-compose.yml
  - Dockerfile
  - pkg\web\frontend\node_modules\@surma\rollup-plugin-off-main-thread\Dockerfile
---

# How do I deploy this project?

## Deployment

### Docker

This project includes Docker configuration.

```bash
docker build -t app .
docker run -p 3000:3000 app
```

### CI/CD

CI/CD pipelines are configured for this project.
Check `.github/workflows/` or equivalent for pipeline configuration.