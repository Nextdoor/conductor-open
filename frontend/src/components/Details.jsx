import React from 'react';
import moment from 'moment';

import Card from 'components/Card';
import TrainComponent from 'components/TrainComponent';
import TitledList from 'components/TitledList';
import {trainProps, requestProps} from 'types/proptypes';

class Details extends TrainComponent {
  render() {
    return (
      <Card className="details-card" header="Details">
        {this.getComponent()}
      </Card>
    );
  }

  getComponent() {
    const requestComponent = this.getRequestComponent();
    if (requestComponent !== null) {
      return requestComponent;
    }

    const {train} = this.props;
    const headCommit = train.commits[train.commits.length - 1];
    const items = [
      ['Branch:', train.branch],
      ['Head:',
        <a className="commit-link" href={headCommit.url}>
          {headCommit.message.split('\n')[0]}
        </a>
      ],
      ['Commits:', train.commits.length],
      ['Engineer:', train.engineer !== null ? train.engineer.name : 'None'],
      ['Created:', moment(train.created_at).fromNow()],
      ['Deployed:', train.deployed_at !== null ? moment(train.deployed_at).fromNow() : 'Not deployed']
    ];

    return <TitledList items={items}/>;
  }
}

Details.propTypes = {
  train: trainProps,
  request: requestProps.isRequired,
};

export default Details;
