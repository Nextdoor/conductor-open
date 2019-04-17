import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/App';

const mapStateToProps = (state) => {
  return {
    needToken: state.token.token === null,
    promptLogin: state.token.promptLogin
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    getToken: () => {
      dispatch(Actions.Token.get());
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
