# Running goEDMS on gokrazy

The aim is to get to a fault tolerant instance so that you can just replace the raspbery pi or SSD WHEN it fails.
For the moment I am using SSD and hopeful for enough time to make it fault tolerant and to develop the backup systems.


# Plan of attack

Preq:
Raspberry 5 with SSD and working gokrazy install with Podman see https://gokrazy.org/packages/docker-containers/ for installing podman (worked well)

- [x]  Got a manual installation working
- [x]  Test with manual connection of local goEDMS to external  
- [ ]  Test this repo  

Then gok add this repo
Configure environmnet variables:

POSTGRES_PASSWORD

## Stage 1 Manual start and manual database creation

- Manully start Postgres with `podman run -e POSTGRES_PASSWORD=1234 -p 5432:5432 --name postgres postgres`
    - You many need to remove and old version with `podman ps -a` then `podman rm postgres` Starting it did not 
    seem to work.
- using postgres user and database goEDMS (failed the first time but worked the second) create the database with dbeaver This is now working.

- attaching and pressing ctrl C exits