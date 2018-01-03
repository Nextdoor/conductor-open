import React from 'react';
import PropTypes from 'prop-types';

import {trainProps, requestProps} from 'types/proptypes';
import Admin from 'containers/Admin';
import Commits from 'containers/Commits';
import Details from 'containers/Details';
import Error from 'components/Error';
import Phases from 'containers/Phases';
import Summary from 'containers/Summary';
import TrainComponent from 'components/TrainComponent';
import TrainHeader from 'containers/TrainHeader';

class Train extends TrainComponent {
  componentWillMount() {
    const {trainId, request, load} = this.props;

    if (request.fetching !== true && request.receivedAt === null) {
      load(trainId);
    }

    this.beginAutoRefresh();
  }

  componentWillReceiveProps(nextProps) {
    this.beginAutoRefresh(nextProps.trainId);
  }

  beginAutoRefresh(trainId) {
    if (this.state && this.state.interval) {
      if (trainId === this.state.intervalTrainId) {
        return;
      }
      clearInterval(this.state.interval);
    }

    const interval = setInterval(function() {
      this.props.load(trainId);
    }.bind(this), 5000);

    this.setState({
      interval: interval,
      intervalTrainId: trainId
    });
  }

  render() {
    const {request} = this.props;

    if (request.fetching !== true && request.receivedAt === null) {
      return null;
    }

    if (request.error !== null) {
      return <Error message={request.error}/>;
    }

    return (
      <div className="grid">
        <div className="column-left">
          <TrainHeader/>
          <Admin/>
          <Summary/>
        </div>
        <div className="column-right">
          <Details/>
          <div className="card queue-card">
            <div className="card-header">
              Queue
            </div>
            <div className="card-divider"/>
          </div>
        </div>
        <div className="column-left">
          <Phases/>
          <Commits/>
        </div>
      </div>
    );
  }
}

Train.propTypes = {
  trainId: PropTypes.string,
  train: trainProps,
  request: requestProps.isRequired,
  load: PropTypes.func.isRequired
};

export default Train;
