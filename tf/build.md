== Pre-deploy - init folder

Create base resources for deployment, one-time, without S3 backend
Also creates VM for deployment

=== Prerequisites

* Folder name
* SSH user and public key
* TG token

=== Result

Created folder with:
* S3 bucket for deployment TF state
* Deployer SA with admin permissions to folder 
* Deployer VM with this SA
* Lockbox secrets with:
    * TG token
    * S3 keys
    * SSH keys

=== Output

* S3 bucket name

== Deploy

=== Prerequisites

Previously created folder with its contents.
1. For deploy:
  * Deployer SA
  * S3 bucket
  * Lockbox with keys to bucket
  * VM for deployment (optional)
2. For service:
  * Folder
  * Lockbox secret with TG token
3. Input
  * Folder name
  * SSH user and public key

=== Output

* Network and subnet
* Service SA with log and docker access
* Docker registry
* Lockbox with SSH creds
* Logging group
* Lockbox with DB password
* DB cluster
* Instance group with COI-service