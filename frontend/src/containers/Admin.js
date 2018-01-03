import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Admin';

const mapStateToProps = (state) => {
  return {
    self: state.self.details,
    train: state.train.details,
    config: state.admin.config,
    fetchConfigRequest: state.admin.requestConfig,
    toggleModeRequest: state.admin.requestToggleMode,
    toggleCloseRequest: state.train.requestToggleClose
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    toggleMode: () => {
      dispatch(Actions.Admin.toggleMode());
    },
    toggleClose: () => {
      dispatch(Actions.Train.toggleClose());
    },
    fetchConfig: () => {
      dispatch(Actions.Admin.fetchConfig());
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
