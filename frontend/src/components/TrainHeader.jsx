import React from 'react';
import PropTypes from 'prop-types';
import {Link} from 'react-router';

import TrainComponent from 'components/TrainComponent';
import {trainProps, requestProps} from 'types/proptypes';

class TrainHeader extends TrainComponent {
  render() {
    const {train, goToTrain} = this.props;
    if (train === null) {
      return (
        <div className="train-header">
          ...
          <div className="train-header-navs">
            <TrainNav arrowType="left" toId={null} goToTrain={goToTrain}/>
            <TrainNav arrowType="right" toId={null} goToTrain={goToTrain}/>
          </div>
        </div>
      );
    }

    let previousTrainId = null;
    let nextTrainId = null;

    if (train.previous_id !== null) {
      previousTrainId = parseInt(train.previous_id, 10);
    }
    if (train.next_id !== null) {
      nextTrainId = parseInt(train.next_id, 10);
    }

    return (
      <div className="train-header">
        Train {train.id}
        <div className="train-header-navs">
          <TrainNav arrowType="left" toId={previousTrainId} goToTrain={goToTrain}/>
          <TrainNav arrowType="right" toId={nextTrainId} goToTrain={goToTrain}/>
        </div>
      </div>
    );
  }
}

TrainHeader.propTypes = {
  request: requestProps.isRequired,
  train: trainProps,
  goToTrain: PropTypes.func.isRequired
};

class TrainNav extends React.Component {
  render() {
    const {arrowType, toId, goToTrain} = this.props;

    let img = null;
    if (toId === null) {
      img = '/images/arrow-' + arrowType + '-disabled.png';
    } else {
      img = '/images/arrow-' + arrowType + '.png';
    }
    const imgTag = <img className="train-header-nav-img" src={img}/>;

    if (toId === null) {
      return (
        <button className="train-header-nav-button" disabled>
          {imgTag}
        </button>
      );
    }

    const uri = '/train/' + toId;
    return (
      <Link to={uri}>
        <button className="button train-header-nav-button" onClick={() => goToTrain(toId)}>
          {imgTag}
        </button>
      </Link>
    );
  }
}

TrainNav.propTypes = {
  toId: PropTypes.number,
  arrowType: PropTypes.string.isRequired,
  goToTrain: PropTypes.func.isRequired
};


export default TrainHeader;
