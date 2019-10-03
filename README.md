# Conductor

**Table of Contents**

- [Introduction](#introduction)
- [Diagrams](#diagrams)
- [Quick Dev Setup](#quick-dev-setup)
- [Terminology](#terminology)
- [Trains](#trains)


## Introduction

Conductor is a Continious Deployment (CD) tool for cherry-picked changes. It is useful in large enterprises with many engineers working on a codebase. Conductor helps make more frequent releases with smaller increments of changelists, so that changes can be landed more frequently into releases. This prevents a large amount of changes from being backed up and reverted due to a slow and cumbersome release cycle.

Central to Conductor's  workflow is the concept of a train. Trains are our smallest unit of deployable code. Commits on a train are deployed to production in a process driven by the engineers themselves. When you cherrypick a change, there are two possibilities:

* Train already exists, your commit "waiting on the platform" with other commits for the next train.
* No train exists, new train created and picks up all of those commits which have been "waiting on the platform".



## Diagrams

Built using [Gravizool](https://github.com/swaggy/gravizool).

### Black Box architecture
![Alt text](https://g.gravizo.com/svg?%0A%20%20%20%20digraph%20G%20{%0A%20%20%20%20%20%20%20%20node[shape="rectangle"]%3B%0A%20%20%20%20%20%20%20%20db%20[label="Postgres%20DB"]%3B%0A%20%20%20%20%20%20%20%20choochoo%20[label="ChooChoo%20Service"]%3B%0A%20%20%20%20%20%20%20%20slack%20[label="Slack"]%3B%0A%20%20%20%20%20%20%20%20jira%20[label="JIRA"]%3B%0A%20%20%20%20%20%20%20%20jenkins%20[label="Jenkins"]%3B%0A%20%20%20%20%20%20%20%20ghe%20[label="Github"]%3B%0A%20%20%20%20%20%20%20%20git_hook%20[label="Git%20Hook"]%3B%0A%20%20%20%20%20%20%20%20web_app%20[label="Web%20App"]%3B%0A%20%20%20%20%20%20%20%20slack_bot%20[label="Slack%20Bot"]%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20db%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20slack%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20jira%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20jenkins%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20ghe%20->%20git_hook%3B%0A%20%20%20%20%20%20%20%20git_hook%20->%20choochoo%3B%0A%20%20%20%20%20%20%20%20web_app%20->%20choochoo%3B%0A%20%20%20%20%20%20%20%20slack_bot%20->%20choochoo%3B%0A%20%20%20%20%20%20%20%20dev%20[label="Some%20Dev%20Box"]%3B%0A%20%20%20%20%20%20%20%20dev%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20dev%20->%20web_app%3B%0A%20%20%20%20%20%20%20%20dev%20->%20slack_bot%3B%0A%20%20%20%20}%0A)

### Full architecture
![Alt text](https://g.gravizo.com/svg?%0A%20%20%20%20digraph%20G%20{%0A%20%20%20%20%20%20%20%20node[shape="rectangle"]%3B%0A%20%20%20%20%20%20%20%20db%20[label="Postgres%20DB"]%3B%0A%20%20%20%20%20%20%20%20slack%20[label="Slack"]%3B%0A%20%20%20%20%20%20%20%20jira%20[label="JIRA"]%3B%0A%20%20%20%20%20%20%20%20jenkins%20[label="Jenkins"]%3B%0A%20%20%20%20%20%20%20%20ghe%20[label="Github"]%3B%0A%20%20%20%20%20%20%20%20git_hook%20[label="Git%20Hook"]%3B%0A%20%20%20%20%20%20%20%20web_app%20[label="Web%20App"]%3B%0A%20%20%20%20%20%20%20%20slack_bot%20[label="Slack%20Bot"]%3B%0A%20%20%20%20%20%20%20%20ghe%20->%20git_hook%3B%0A%20%20%20%20%20%20%20%20dev%20[label="Some%20Dev%20Box"]%3B%0A%20%20%20%20%20%20%20%20dev%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20dev%20->%20web_app%3B%0A%20%20%20%20%20%20%20%20dev%20->%20slack_bot%3B%0A%20%20%20%20%20%20%20%20api%20[label="API"]%3B%0A%20%20%20%20%20%20%20%20data%20[label="Data"]%3B%0A%20%20%20%20%20%20%20%20build%20[label="Build"]%3B%0A%20%20%20%20%20%20%20%20code%20[label="Code"]%3B%0A%20%20%20%20%20%20%20%20deploy%20[label="Deploy"]%3B%0A%20%20%20%20%20%20%20%20messaging%20[label="Messaging"]%3B%0A%20%20%20%20%20%20%20%20verification%20[label="Verification"]%3B%0A%20%20%20%20%20%20%20%20data%20->%20db%3B%0A%20%20%20%20%20%20%20%20build%20->%20jenkins%3B%0A%20%20%20%20%20%20%20%20code%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20deploy%20->%20jenkins%3B%0A%20%20%20%20%20%20%20%20messaging%20->%20slack%3B%0A%20%20%20%20%20%20%20%20verification%20->%20jira%3B%0A%20%20%20%20%20%20%20%20git_hook%20->%20api%3B%0A%20%20%20%20%20%20%20%20web_app%20->%20api%3B%0A%20%20%20%20%20%20%20%20slack_bot%20->%20api%3B%0A%20%20%20%20%20%20%20%20api%20->%20data%3B%0A%20%20%20%20%20%20%20%20api%20->%20build%3B%0A%20%20%20%20%20%20%20%20api%20->%20code%3B%0A%20%20%20%20%20%20%20%20api%20->%20deploy%3B%0A%20%20%20%20%20%20%20%20api%20->%20messaging%3B%0A%20%20%20%20%20%20%20%20api%20->%20verification%3B%0A%20%20%20%20}%0A)

## Quick Dev Setup

### Download Tools

-If missing, install latest GoLang (v1.13 or above), by downloading and running [official installer](https://golang.org/dl/#featured)

-If missing, install latest nodejs by downloading and running [official installer](https://nodejs.org/en/download/)

-If missing, install latest docker client, by signing up and [downloading client](https://docs.docker.com/docker-for-mac/install/)

-If missing nginx (optional; for local machine setup only), install it via `brew install nginx` or compile from [source](http://nginx.org/en/download.html)

NOTE: make sure docker is up and running (either `docker up` or check if you see client running on top bar of your mac OS screen), on your machine before proceeding to running setup scripts.

### Download Source Code

-Download project in the given path relative to GO path:

```
cd $GOPATH/src
mkdir conductor
cd conductor

git clone https://github.com/Nextdoor/conductor.git
```

### Simple Deployment of Conductor on Dev Environment

1. On your terminal `cd $GOPATH/src/conductor/conductor`
2. Create a client and OAuth token from you git login by going to `https://github.com/settings/developers`. Create a `New OAuth App`. Set Application Name to `conductor`.  Homepage URL to `https://github.com/Nextdoor/conductor`.  And Authorized callback URL to your local dev URL for now `http://localhost/api/auth/login`
3. Now, set the client id created in step above, in the `OAUTH_PAYLOAD` variable in `frontend/envfile` of your downloaded source code.
4. To create a conductor setup on a docker container , run `chmod +x ./dockerSetup.sh`, followed by `./dockerSetup.sh`
5. To create a conductor setup on native mac , run `chmod +x ./nativeMacSetup.sh`, followed by `./nativeMacSetup.sh`
6. In either case, your docker service is now accessible through your browser on `localhost:80`
7. While development, if you wanted to quickly swap your code into the local mac deployment, you could run the script with frontend/ backend flags. (Note: for frontend, you need to have backend running in a different terminal)

```
./nativeMacSetup.sh --frontend
./nativeMacSetup.sh --backend
```

NOTE: This is without integrations with GitHub, Slack, Jenkins, Jira. These are specific features that you can setup on a need-for basis on your dev environment, by generating real access keys with these systems and setting them into the `envfile`. Otherwise we use mocked simulation of these tools in dev


### Debugging Instructions

For the native environment debugging (which is generally faster than developing on docker), you can use VSCode, and add a debug configuration to you launch.json as below.
When the debugger line goes from blue to orange it mean it attached successfully, and should now hit your breakpoints. NOTE: VSCode doesn’t appear to support remote debugging for Go language yet.

```
"version": "0.2.0",
    "configurations": [

        {
            "name": "Local deployment",
            "mode": "local",
	    "type": "go",
            "request": "attach",
            "processId": your-local-conductor-process-id
        }

    ]
```

OR

You can attach a debugger either on your native machine terminal (of IDE), or within the shell of the conductor docker container.
The instruction in either case is 

1) Download delve debugger:
```
go get -u github.com/go-delve/delve/cmd/dlv
```

2) Get the process id of the running conductor process, and run the delve command

```
ps -ef | grep conductor
dlv attach <your-local-conductor-process-id> --headless=true --listen=localhost --api-version=2 --wd=$GOPATH/src/conductor/conductor
```

### Few Other Useful Commands

1) To attach to the shell of a docker image, do `docker ps`, followed my `docker exec -t <container-id> bash`. Or get the Docker plugin on VSCode, and right click and attach shell on the running containers panel.

2) If changes on your local code are not reflecting on your docker container deployment. Delete old conductor docker volume and images (VSCode Docker plugin is very helpful for seeing this visually), and run the ./dockerSetup.sh command again.

2) If changes on your local code are not reflecting on your native mac conductor deployment. `rm -rf ~/app/conductor`, and run the ./nativeMacSetup.sh command again.

3) In local deployment to restart your nginx server, try `sudo nginx -s stop` and `sudo nginx -c $HOME/app/nginx.conf -p $HOME/app/`

4) To see if an existing process is running on a port try  `lsof -i :<port-number>`

5) To kill and debugger process which would prevent re-running conductor, simply get process id by `ps -ef | grep conductor` and `kill -9` all of the listed results.




## Terminology

* train  - A group of commits to be deployed
    * engineer - The person in charge of the train, who gets the honor of starting the deploy
    * train derailment - Destroying the currently running train
    * auto-time - Time when trains are automatically created
* build  - A build for that group of commits
* ticket - A verification that needs to be completed for the deploy to begin
* deploy - Process of releasing these commits to production

## Trains

### Summary

1. Normally, trains are auto-created. Upon creation, a train will include all the queued commits on the branch.
2. If a problem is found with a commit:
   1. The commit is reverted and the revert is put on the queue
   2. The train is derailed
   3. A new train is created as the addition of the old train and the queue. Tickets transfer.
3. Trains will not be auto-created outside of configured auto-time or when `conductor` is in manual mode.
4. Trains can be manually created only when `conductor` is in manual mode.
5. When an emergency fix needs to be done:
   1. `conductor` is put into manual mode
   2. The current train, if any, is derailed
   3. The branch is reset to before the old train, removing the queue as well
   4. The emergency commits are pushed
   5. A new train is created manually

---

A train begins whenever possible.

There can only be one train going at a time.

As soon as there is no train going, `conductor` will start the next train as soon as possible.

Commits are pushed through version control, which `conductor` watches.

Here's an example of how this looks:

    Train going.
    Four additional commits pushed.
    Train deploys.
    New train starts with those four commits.

Here's another example, where there are no queued commits after a train is finished.

    Train going.
    No additional commits pushed.
    Train deploys.
    Waits...
    Commit pushed.
    New train starts with that one commit.

There are some additional situations that must be considered.

`conductor` is configurable to stop ordinary release procedure after certain time periods.

    Train going.
    Two additional commits pushed.
    Train deploys.
    New train doesn't start because it's not within configured auto-time.
    Waits...
    Five additional commits pushed.
    Back within configured auto-time.
    New train starts with those seven commits.

This stopped behavior can be interrupted when necessary, like in any situations that require an emergency release.

    Train going.
    Two additional commits pushed.
    Train deploys.
    New train doesn't start because it's not within configured auto-time.
    One additional commit pushed.
    New train force started with those three additional commits.

In these situations, committers of queued commits may be unable to verify their changes.

Either way, those commits add unnecessary risk.

There must be a way to limit commits when necessary: pull them off.

    Train going.
    Two additional commits pushed.
    Train deploys.
    New train doesn't start because it's not within configured auto-time.
    One additional commit pushed.
    The two original commits are reverted.
    New train force started with only the latest commit.

What if there was an important situation during auto-time?

We can do the same thing there.

    Train going.
    Emergency situation!
    Put `conductor` in manual mode and cancel the current train.
    Revert unrelated commits.
    Hand-pick commits for the new train.

This looks good, but has some issues.

    A) There are race conditions between reverts, train creation, and people pushing their changes.
    B) When things are okay, the reverts themselves will be reverted, leading to lots of "revert revert" commits.

We can do better.

Let's imagine this is the state of the branch, where each letter is a commit.

The train is defined between '[' and ']'.

    ... [a b c] d e

Now, an emergency fix needs to come in, call it 'f', and all of 'a', 'b', 'c', 'd', and 'e' should be removed.

How should this be done?

Well, cleanest would be to reset the branch to right before the train, before 'a'.

So, in these emergency situations, let's do that. A special "remove train and queue" operation.

The fact that `conductor` is in manual mode can be used to warn people trying to commit.

What about non-emergency situations, where something is found wrong in the staging environment?

    Train going.
    Commits 'a', 'b', and 'c' on the train.
    Commit 'd' added to queue.
    Problem found with commit 'b'.
    ???

At this point, there are two options.

    A) Extend the train to include the revert commit and the queue.
    B) Remove d from the branch with a force push and put the revert commit on.

Option B) may lead to delays in getting queued code out.

    [a b c d e f g] h i j k l m n o
    ... 10m of build & verification
    Problem with e - delete the queue
    [a b c d e f g !e] h i j k l m n o p q r s
    ... 10m of build & verification
    Problem with g - delete the queue.
    ...
    And the queue would keep building and building.

The queue shouldn't be let to grow like that.

Instead, we extend the train to include the previous train and the queue.

Keep in mind that 'a' and 'c' can still be verified while the new train is building.

Tickets from the original train are carried over.

Then this larger train goes out all at once.
