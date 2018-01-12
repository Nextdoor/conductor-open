# Conductor

**Table of Contents**

- [Diagrams](#diagrams)
- [Terminology](#terminology)
- [Trains](#trains)

## Diagrams

Built using [Gravizool](https://github.com/swaggy/gravizool).

### Black Box architecture
![Alt text](https://g.gravizo.com/svg?%0A%20%20%20%20digraph%20G%20{%0A%20%20%20%20%20%20%20%20node[shape="rectangle"]%3B%0A%20%20%20%20%20%20%20%20db%20[label="Postgres%20DB"]%3B%0A%20%20%20%20%20%20%20%20choochoo%20[label="ChooChoo%20Service"]%3B%0A%20%20%20%20%20%20%20%20slack%20[label="Slack"]%3B%0A%20%20%20%20%20%20%20%20jira%20[label="JIRA"]%3B%0A%20%20%20%20%20%20%20%20jenkins%20[label="Jenkins"]%3B%0A%20%20%20%20%20%20%20%20ghe%20[label="Github"]%3B%0A%20%20%20%20%20%20%20%20git_hook%20[label="Git%20Hook"]%3B%0A%20%20%20%20%20%20%20%20web_app%20[label="Web%20App"]%3B%0A%20%20%20%20%20%20%20%20slack_bot%20[label="Slack%20Bot"]%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20db%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20slack%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20jira%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20jenkins%3B%0A%20%20%20%20%20%20%20%20choochoo%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20ghe%20->%20git_hook%3B%0A%20%20%20%20%20%20%20%20git_hook%20->%20choochoo%3B%0A%20%20%20%20%20%20%20%20web_app%20->%20choochoo%3B%0A%20%20%20%20%20%20%20%20slack_bot%20->%20choochoo%3B%0A%20%20%20%20%20%20%20%20dev%20[label="Some%20Dev%20Box"]%3B%0A%20%20%20%20%20%20%20%20dev%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20dev%20->%20web_app%3B%0A%20%20%20%20%20%20%20%20dev%20->%20slack_bot%3B%0A%20%20%20%20}%0A)

### Full architecture
![Alt text](https://g.gravizo.com/svg?%0A%20%20%20%20digraph%20G%20{%0A%20%20%20%20%20%20%20%20node[shape="rectangle"]%3B%0A%20%20%20%20%20%20%20%20db%20[label="Postgres%20DB"]%3B%0A%20%20%20%20%20%20%20%20slack%20[label="Slack"]%3B%0A%20%20%20%20%20%20%20%20jira%20[label="JIRA"]%3B%0A%20%20%20%20%20%20%20%20jenkins%20[label="Jenkins"]%3B%0A%20%20%20%20%20%20%20%20ghe%20[label="Github"]%3B%0A%20%20%20%20%20%20%20%20git_hook%20[label="Git%20Hook"]%3B%0A%20%20%20%20%20%20%20%20web_app%20[label="Web%20App"]%3B%0A%20%20%20%20%20%20%20%20slack_bot%20[label="Slack%20Bot"]%3B%0A%20%20%20%20%20%20%20%20ghe%20->%20git_hook%3B%0A%20%20%20%20%20%20%20%20dev%20[label="Some%20Dev%20Box"]%3B%0A%20%20%20%20%20%20%20%20dev%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20dev%20->%20web_app%3B%0A%20%20%20%20%20%20%20%20dev%20->%20slack_bot%3B%0A%20%20%20%20%20%20%20%20api%20[label="API"]%3B%0A%20%20%20%20%20%20%20%20data%20[label="Data"]%3B%0A%20%20%20%20%20%20%20%20build%20[label="Build"]%3B%0A%20%20%20%20%20%20%20%20code%20[label="Code"]%3B%0A%20%20%20%20%20%20%20%20deploy%20[label="Deploy"]%3B%0A%20%20%20%20%20%20%20%20messaging%20[label="Messaging"]%3B%0A%20%20%20%20%20%20%20%20verification%20[label="Verification"]%3B%0A%20%20%20%20%20%20%20%20data%20->%20db%3B%0A%20%20%20%20%20%20%20%20build%20->%20jenkins%3B%0A%20%20%20%20%20%20%20%20code%20->%20ghe%3B%0A%20%20%20%20%20%20%20%20deploy%20->%20jenkins%3B%0A%20%20%20%20%20%20%20%20messaging%20->%20slack%3B%0A%20%20%20%20%20%20%20%20verification%20->%20jira%3B%0A%20%20%20%20%20%20%20%20git_hook%20->%20api%3B%0A%20%20%20%20%20%20%20%20web_app%20->%20api%3B%0A%20%20%20%20%20%20%20%20slack_bot%20->%20api%3B%0A%20%20%20%20%20%20%20%20api%20->%20data%3B%0A%20%20%20%20%20%20%20%20api%20->%20build%3B%0A%20%20%20%20%20%20%20%20api%20->%20code%3B%0A%20%20%20%20%20%20%20%20api%20->%20deploy%3B%0A%20%20%20%20%20%20%20%20api%20->%20messaging%3B%0A%20%20%20%20%20%20%20%20api%20->%20verification%3B%0A%20%20%20%20}%0A)

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
