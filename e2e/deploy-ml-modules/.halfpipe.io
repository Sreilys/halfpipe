team: test
pipeline: test
repo:
  watched_paths:
  - e2e/deploy-ml-modules

tasks:
- type: deploy-ml-modules
  name: Deploy ml-modules artifact
  ml_modules_version: "2.1425"
  app_name: my-app
  app_version: v1
  targets:
  - ml.dev.springer-sbm.com
  - ml.qa1.springer-sbm.com
