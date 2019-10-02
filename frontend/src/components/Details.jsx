import React from 'react';
import moment from 'moment';

import ApiButton from 'components/ApiButton';
import Card from 'components/Card';
import TrainComponent from 'components/TrainComponent';
import TitledList from 'components/TitledList';
import {trainProps, requestProps} from 'types/proptypes';

class Details extends TrainComponent {

  render() {
    return (
      <Card className="details-card" header="Details">
        {this.getComponent()}
        {this.claimEngineerButton()}
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

  claimEngineerButton() {
    return (
      <ApiButton
        modalProps={{
          title: 'Confirm Claiming Engineer Role',
          body: (
            <div>
              Thanks for volunteering to guide this train.
              <br/><br/>
              As train engineer, that all commits have been verified (if needed), and that the build together passes, so that the train can proceed to deploy.
              <br/><br/>
              Also please toubleshoot (or reach an admin), to take the train to succesful train deployment and closing.
            </div>
          )
        }}
        //onClick={() => this.props.restartJob(this.props.train.id, 'deploy')}
        request={this.props.request}
        className="button-claim">
        Claim Engineer Role
      </ApiButton>
    );
  }
}

Details.propTypes = {
  train: trainProps,
  request: requestProps.isRequired,
};

export default Details;
