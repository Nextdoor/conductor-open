import React from 'react';
import PropTypes from 'prop-types';

import ApiButton from 'components/ApiButton';
import Card from 'components/Card';
import TitledList from 'components/TitledList';
import TrainComponent from 'components/TrainComponent';
import {trainProps, requestProps} from 'types/proptypes';
import {Phases} from 'types/train';

import moment from 'moment';

class Summary extends TrainComponent {
  render() {
    return (
      <Card className="summary-card" header="Summary">
        {this.getComponent()}
      </Card>
    );
  }

  getComponent() {
    const requestComponent = this.getRequestComponent();
    if (requestComponent !== null) {
      return requestComponent;
    }

    const train = this.props.train;

    // TODO: Figure out how to surface that the current user is on the train / their specific commit state.

    let summary = null;
    if (train.done) {
      summary = this.doneSummary();
    } else {
      switch (train.active_phase) {
        default:
        case Phases.Delivery:
          summary = this.deliverySummary();
          break;
        case Phases.Verification:
          summary = this.verificationSummary();
          break;
        case Phases.Deploy:
          summary = this.deploySummary();
          break;
      }
    }

    return (
      <div>
        <div className="summary-details">
          {summary}
        </div>
        <div className="summary-buttons">
          {this.extendButton()}
          {this.blockButton()}
          {this.unblockButton()}
          {this.cancelButton()}
          {this.rollbackButton()}
        </div>
      </div>
    );
  }

  deliverySummary() {
    const items = [
      ['Status:', (
        <span className="summary-status-text">
          {this.isExtended() ? 'Train extended. Delivering new changes to staging.' : 'Delivering new changes to staging'}
        </span>
      )],
      ['Jobs:', this.getJobStatusLine(this.props.train.active_phases.delivery.jobs)]
    ];

    const peopleNames = this.getPeopleInDelivery();
    if (peopleNames.size > 0) {
      items.push(
        ['Changes by:', [...peopleNames]
          .reduce((prev, curr) => [prev, ', ', curr])]);
    }

    return <TitledList items={items}/>;
  }

  verificationSummary() {
    let not_deployable_reason = 'Waiting for verification';
    if (this.props.train.not_deployable_reason !== null) {
      not_deployable_reason = this.props.train.not_deployable_reason;
    }
    const items = [
      ['Status:', <span className="summary-status-text">{not_deployable_reason}</span>],
      ['Jobs:', this.getJobStatusLine(this.props.train.active_phases.verification.jobs)],
      ['Tickets:', this.getTicketStatusLine(this.props.train.tickets)],
    ];

    const peopleNames = this.getPeopleWithOpenTickets();
    if (peopleNames.size > 0) {
      items.push(
        ['Open tickets:', [...peopleNames]
          .reduce((prev, curr) => [prev, ', ', curr])]);
    }

    return <TitledList items={items}/>;
  }

  deploySummary() {
    const items = [
      ['Status:', <span className="summary-status-text">Deploying to production.</span>],
      ['Jobs:', this.getJobStatusLine(this.props.train.active_phases.deploy.jobs)],
    ];

    return <TitledList items={items}/>;
  }

  doneSummary() {
    const deployed_at = this.props.train.deployed_at;
    const cancelled_at = this.props.train.cancelled_at;

    let text = '';
    if (deployed_at !== null) {
      text = <span>This train was deployed to production on {moment(deployed_at).format('llll')}.</span>;
    } else if (cancelled_at !== null) {
      text = <span>This train was cancelled on {moment(cancelled_at).format('llll')}.</span>;
    } else {
      text = <span>This train was never deployed to production.</span>;
    }

    const items = [
      ['Status:', <span className="summary-status-text">{text}</span>],
    ];

    return <TitledList items={items}/>;
  }

  isExtended() {
    return this.props.train.all_phase_groups.length > 1;
  }

  getPeopleInDelivery() {
    const train = this.props.train;

    const peopleNames = new Set();
    const commitList = this.getCommitsBetweenSHAs(train.active_phases.head_sha, train.last_delivered_sha);

    for (let i = 0; i < commitList.length; i++) {
      const commit = commitList[i];
      if (!(peopleNames.has(commit.author_name))) {
        peopleNames.add(commit.author_name);
      }
    }

    return peopleNames;
  }

  // Return includes head but not tail.
  // If tail is null, this includes every commit up to the head sha.
  getCommitsBetweenSHAs(headSHA, tailSHA) {
    const commits = this.props.train.commits;

    let inBetween = false;
    if (tailSHA === null) {
      inBetween = true;
    }

    const commitList = [];
    for (let i = 0; i < commits.length; i++) {
      const commit = commits[i];
      if (inBetween) {
        commitList.push(commit);
      }
      if (commit.sha === tailSHA) {
        // After this commit, start adding all commits until we reach the current head sha.
        inBetween = true;
      } else if (commit.sha === headSHA) {
        // Done.
        break;
      }
    }

    return commitList;
  }

  getPeopleWithOpenTickets() {
    const train = this.props.train;

    const peopleNames = new Set();

    for (let i = 0; i < train.tickets.length; i++) {
      const ticket = train.tickets[i];
      if (ticket.closed_at !== null || ticket.deleted_at !== null) {
        continue;
      }
      if (!(peopleNames.has(ticket.assignee_name))) {
        peopleNames.add(ticket.assignee_name);
      }
    }

    return peopleNames;
  }

  getJobStatus(jobs) {
    return {
      completedCount: jobs.filter((job) => job.completed_at !== null && job.result === 0).length,
      totalCount: jobs.length
    };
  }

  getJobStatusLine(jobs) {
    const jobStatus = this.getJobStatus(jobs);

    return <span>{jobStatus.completedCount}/{jobStatus.totalCount}</span>;
  }

  getTicketStatus(tickets) {
    return {
      completedCount: tickets.filter((ticket) => ticket.closed_at !== null || ticket.deleted_at !== null).length,
      totalCount: tickets.length
    };
  }

  getTicketStatusLine(tickets) {
    const ticketStatus = this.getTicketStatus(tickets);

    return <span>{ticketStatus.completedCount}/{ticketStatus.totalCount}</span>;
  }

  extendButton() {
    if (!this.isTrainExtendable(this.props.train)) {
      return;
    }
    return (
      <ApiButton
        modalProps={{
          title: 'Confirm train extension',
          body: (
            <div>
              Any new commits on the branch will be added to the train.
              <br/>
              This will delay the train and make it larger.
              <br/><br/>
              There should be a good reason to extend, such as:
              <ul>
                <li>A necessary fix for the current train.</li>
                <li>A necessary fix for a critical bug in production.</li>
              </ul>
            </div>
          )
        }}
        onClick={() => this.props.extendTrain(this.props.train.id)}
        request={this.props.requestExtend}
        className="js-extend-train">
        Extend
      </ApiButton>
    );
  }

  blockButton() {
    if (!this.isTrainBlockable(this.props.train)) {
      return;
    }
    return (
      <ApiButton
        modalProps={{
          title: 'Confirm train block',
          body: (
            <div>
              This will prevent the train from deploying.
              <br/><br/>
              There should be a good reason to block, such as:
              <ul>
                <li>Something going on or wrong with production that deserves stopping releases.</li>
                <li>An issue discovered for this train that isn't gated by a verification ticket.</li>
              </ul>
            </div>
          )
        }}
        onClick={() => this.props.blockTrain(this.props.train.id)}
        request={this.props.requestBlock}
        className="button-warning js-block-train">
        Block
      </ApiButton>
    );
  }

  unblockButton() {
    if (!this.isTrainUnblockable(this.props.train)) {
      return;
    }
    return (
      <ApiButton
        modalProps={{
          title: 'Confirm train unblock',
          body: (
            <div>
              The train will be made deployable again.
              <br/><br/>
              Unblocking should only be done if the cause for blocking it has been resolved.
            </div>
          )
        }}
        onClick={() => this.props.unblockTrain(this.props.train.id)}
        request={this.props.requestUnblock}
        className="button-warning js-unblock-train">
        Unblock
      </ApiButton>
    );
  }

  cancelButton() {
    if (!this.isTrainCancellable(this.props.train)) {
      return;
    }
    return (
      <ApiButton
        modalProps={{
          title: 'Confirm train cancel',
          body: (
            <div>
              The train will be cancelled.
              <br/><br/>
              This should be done if the deploy was failed or was aborted and is not going to be rerun.
            </div>
          )
        }}
        onClick={() => this.props.cancelTrain(this.props.train.id)}
        request={this.props.requestCancel}
        className="button-warning js-cancel-train">
        Cancel
      </ApiButton>
    );
  }

  rollbackButton() {
    if (!this.isTrainRollbackable(this.props.train)) {
      return;
    }
    return (
      <ApiButton
        modalProps={{
          title: 'Confirm train rollback',
          body: (
            <div>
              This will initiate a rollback to the selected train.
              <br/><br/>
              Should be used if there's a production issue and the latest train is suspected.
            </div>
          )
        }}
        onClick={() => this.props.rollbackToTrain(this.props.train.id)}
        request={this.props.requestRollback}
        className="button-warning button-wide js-rollback-train">
        Rollback To
      </ApiButton>
    );
  }

  isTrainExtendable(train) {
    return this.isUser() && !train.done && train.active_phases.deploy.started_at === null;
  }

  isTrainBlockable(train) {
    return this.isUser() && !train.done && train.active_phases.deploy.started_at === null && train.blocked === false;
  }

  isTrainUnblockable(train) {
    return this.isUser() && !train.done && train.active_phases.deploy.started_at === null && train.blocked === true;
  }

  isTrainCancellable(train) {
    return this.isUser() && !train.done;
  }

  isTrainRollbackable(train) {
    return this.isUser() && train.can_rollback;
  }

  isUser() {
    return this.props.self.is_user || this.props.self.is_admin;
  }
}

Summary.propTypes = {
  self: PropTypes.shape({
    name: PropTypes.string.isRequired,
    email: PropTypes.string.isRequired,
    is_user: PropTypes.bool.isRequired,
    is_admin: PropTypes.bool.isRequired,
  }),
  train: trainProps,
  request: requestProps.isRequired,
  extendTrain: PropTypes.func.isRequired,
  requestExtend: requestProps.isRequired,
  blockTrain: PropTypes.func.isRequired,
  requestBlock: requestProps.isRequired,
  unblockTrain: PropTypes.func.isRequired,
  requestUnblock: requestProps.isRequired,
  cancelTrain: PropTypes.func.isRequired,
  requestCancel: requestProps.isRequired,
  rollbackToTrain: PropTypes.func.isRequired,
  requestRollback: requestProps.isRequired,
};

export default Summary;
