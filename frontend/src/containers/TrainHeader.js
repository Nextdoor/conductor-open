import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/TrainHeader';

const mapStateToProps = (state) => {
  return {
    train: state.train.details,
    request: state.train.request
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    goToTrain: (trainId) => {
      dispatch(Actions.Train.goToTrain(trainId));
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
