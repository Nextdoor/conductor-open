import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Search';

const mapStateToProps = (state) => {
  return {
    details: state.search.details,
    request: state.search.request
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    search: (params) => {
      dispatch(Actions.Search.fetch(params));
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
