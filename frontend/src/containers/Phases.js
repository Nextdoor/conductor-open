import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Phases';

const mapStateToProps = (state) => {
  return {
    self : state.self,
    train: state.train.details,
    request: state.train.request,
    requestRestart: state.train.requestRestart,
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    restartJob: (trainId, phaseName) => {
      dispatch(Actions.Train.restart(trainId, phaseName));
    }
  };
};


export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);

