---
layout: documentation
title: Global Environmental Variables list
date: 2017-08-17
doctype: general
---

{:.table}
Variable | Type | Description
--- | --- | ---
KUBICORN_STATE_STORE | string | The state store type to use for the cluster
KUBICORN_STATE_STORE_PATH | string | The state store path to use
KUBICORN_GIT_CONFIG | string | The git remote ulr to use
KUBICORN_NAME | string | The name of the cluster to use
KUBICORN_PROFILE | string | The profile name to create new clusters APIs with
KUBICORN_SET | string | Set custom property for the cluster
KUBICORN_TRUECOLOR | bool | Always run `kubicorn` with lolgopher truecolor
KUBICORN_ENVIRONMENT | string | If it's set to `LOCAL`, `kubicorn` will use bootstrap local bootstrap scripts instead of remote ones. 
KUBICORN_OUTPUT | string | Set output format for command
KUBICORN_FORCE_DELETE_KEY | bool | Force delete key for AWS or Packet
KUBICORN_FORCE_LOCAL_BOOTSTRAP | bool | Force read bootstrap scripts from local dir / bootstrap
--- | --- | ---
KUBICORN_S3_ACCESS_KEY | string | Access key for S3-compatible object storage
KUBICORN_S3_SECRET_KEY | string | Secret key for S3-compatible object storage
KUBICORN_S3_ENDPOINT | string | Endpoint URL of S3-compatible object storage
KUBICORN_S3_SSL | bool | Use SSL to access S3-compatible object storage
KUBICORN_S3_BUCKET | string | Name of the S3-compatible bucket
--- | --- | ---
AWS_PROFILE | string | The name of the Amazon profile stored in `~/.aws/credentials`
--- | --- | ---
AWS_ACCESS_KEY_ID | string | The AWS access key to use with AWS profiles - Optional, see [AWS Walkthrough](http://kubicorn.io/documentation/aws-walkthrough.html)
AWS_SECRET_ACCESS_KEY | string | The AWS secret to use with AWS profiles - Optional, see [AWS Walkthrough](http://kubicorn.io/documentation/aws-walkthrough.html)
--- | --- | ---
DIGITALOCEAN_ACCESS_TOKEN | string | The DigitalOcean access token used to authenticate with the API
--- | --- | ---
GOOGLE_APPLICATION_CREDENTIALS | string | The location of the Google service account key file
--- | --- | ---
OS_AUTH_URL | string | The URL of the Openstack Identity service
OS_USERNAME | string | The name of the Openstack user
OS_PASSWORD | string | The password of the Openstack user
OS_TENANT_ID | string | The identifier of the Openstack tenant - either this or OS_TENANT_NAME should be set
OS_TENANT_NAME | string | The name of the Openstack tenant - either this or OS_TENANT_ID should be set
OS_DOMAIN_ID | string | The identifier of the Openstack domain - (identity v3) either this or OS_DOMAIN_NAME should be set
OS_DOMAIN_NAME | string | The name of the Openstack domain - (identity v3) either this or OS_DOMAIN_ID should be set
--- | --- | ---
PACKET_APITOKEN | string | The Packet API token used to authenticate with the API
KUBICORN_FORCE_DELETE_PROJECT | bool | Force delete Packet project
