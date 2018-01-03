import React from 'react';

import Card from 'components/Card';
import TrainComponent from 'components/TrainComponent';
import {trainProps, requestProps} from 'types/proptypes';

class Commits extends TrainComponent {
  render() {
    return (
      <Card header="Commits" className="commit-details">
        {this.getComponent()}
      </Card>
    );
  }

  getComponent() {
    const requestComponent = this.getRequestComponent();
    if (requestComponent !== null) {
      return requestComponent;
    }

    return (
      <ul className="commit-groups">
        {this.getGroups().map((group, i) =>
          <li key={i} className="commit-group">
            <span className="commit-group-author">{group[0]}</span>
            {group[1].map((commit, j) =>
              <a key={j} className="commit-link" href={commit.url}>{commit.message}</a>
            )}
            {group[2].map((ticket, j) =>
              <a key={j} className={'ticket-link' + (ticket.done ? ' done' : '')} href={ticket.url}>{ticket.key}</a>
            )}
          </li>
        )}
      </ul>
    );
  }

  getGroups() {
    const {train} = this.props;
    const groupsByEmail = {};

    for (let i = 0; i < train.commits.length; i++) {
      const commit = train.commits[i];
      if (!(commit.author_email in groupsByEmail)) {
        groupsByEmail[commit.author_email] = {
          author_name: commit.author_name,
          commits: [],
          tickets: []
        };
      }
      groupsByEmail[commit.author_email].commits.push({
        message: commit.message.split('\n')[0],
        url: commit.url
      });
    }

    for (let i = 0; i < train.tickets.length; i++) {
      const ticket = train.tickets[i];
      if (!(ticket.assignee_email in groupsByEmail)) {
        groupsByEmail[ticket.assignee_email] = {
          author_name: ticket.assignee_name,
          commits: [],
          tickets: []
        };
      }
      groupsByEmail[ticket.assignee_email].tickets.push({
        key: ticket.key,
        url: ticket.url,
        done: ticket.closed_at !== null || ticket.deleted_at !== null
      });
    }

    const groups = [];

    const authors = Object.keys(groupsByEmail);
    for (let i = 0; i < authors.length; i++) {
      const group = groupsByEmail[authors[i]];
      groups.push([group.author_name, group.commits, group.tickets]);
    }

    return groups;
  }
}

Commits.PropTypes = {
  train: trainProps,
  request: requestProps.isRequired
};

export default Commits;
