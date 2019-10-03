import React from 'react';
import moment from 'moment';

import ApiButton from 'components/ApiButton';
import Card from 'components/Card';
import TrainComponent from 'components/TrainComponent';
import TitledList from 'components/TitledList';
import {trainProps, requestProps} from 'types/proptypes';

class Details extends TrainComponent {

  render() {
    const {train} = this.props;
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
    const trainEngineer = train.engineer !== null ? train.engineer.name : 'None'
    const headCommit = train.commits[train.commits.length - 1];
    const items = [
      ['Branch:', train.branch],
      ['Head:',
        <a className="commit-link" href={headCommit.url}>
          {headCommit.message.split('\n')[0]}
        </a>
      ],
      ['Commits:', train.commits.length],
      ['Engineer:', trainEngineer],
      ['Created:', moment(train.created_at).fromNow()],
      ['Deployed:', train.deployed_at !== null ? moment(train.deployed_at).fromNow() : 'Not deployed']
    ];

    return (
      <span>
        <TitledList items={items}/>
        {this.claimEngineerButton(trainEngineer)}
      </span>
    )
  }

  claimEngineerButton(trainEngineer) {    
    
   let message = {
    title: 'Become the engineer for this train',
    body: (
      <div>
        By clicking confirm, you will replace {trainEngineer} as the engineer for this train. 
        <br/><br/>
        Thank you for keeping our trains on schedule!
      </div>
    )
  }
   if (!trainEngineer){
     message = {
      title: 'Become the engineer for this train',
      body: (
        <div>
          By clicking confirm, you will become the engineer for this train. 
          <br/><br/>
          Thank you for keeping our trains on schedule!
        </div>
      )
    }
   }
   if(!this.props.train.closed)
    {
      return (
        <ApiButton
          modalProps={message}
          onClick={() => this.props.changeEngineer(this.props.train.id)}
          request={this.props.request}
          className="button-claim">
          Claim Engineer Role
        </ApiButton>
      );
    }
  }
}

Details.propTypes = {
  train: trainProps,
  request: requestProps.isRequired,
};

export default Details;
