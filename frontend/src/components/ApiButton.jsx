import React from 'react';
import {Modal} from 'react-overlays';
import ReactDOM from 'react-dom';
import PropTypes from 'prop-types';
import {requestProps} from 'types/proptypes';

import Loading from 'components/Loading';

class ApiButton extends React.Component {
  constructor(props) {
    super(props);
    this.state = {inModal: false};
  }

  render() {
    if (this.props.request.fetching === true) {
      return (
        <button className={this.props.className}
          disabled>
          <Loading/>
        </button>
      );
    }

    return (
      <button className={this.props.className}
        onClick={this.clicked.bind(this)}
        disabled={this.props.enabled}>
        {this.props.children}
        {this.state.inModal && <ApiButtonModal
          onCancel={this.modalCancel.bind(this)}
          onConfirm={this.modalConfirm.bind(this)}
          {...this.props.modalProps}/>}
      </button>
    );
  }

  clicked() {
    ReactDOM.findDOMNode(this).blur();

    if (this.props.modalProps && !this.state.inModal) {
      this.setState({inModal: true});
    } else {
      this.props.onClick();
    }
  }

  modalCancel() {
    this.setState({inModal: false});
  }

  modalConfirm() {
    this.setState({inModal: false});
    this.props.onClick();
  }
}

ApiButton.propTypes = {
  enabled: PropTypes.bool,
  className: PropTypes.string,
  modalProps: PropTypes.shape({
    title: PropTypes.node,
    body: PropTypes.node,
    cancel: PropTypes.node,
    confirm: PropTypes.node,
  }),
  onClick: PropTypes.func.isRequired,
  request: requestProps.isRequired,
  children: PropTypes.node,
};

class ApiButtonModal extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      title: this.props.title || 'title not implemented',
      body: this.props.body || 'body not implemented',
      cancel: this.props.cancel || 'Cancel',
      confirm: this.props.confirm || 'Confirm',
    };
  }

  render() {
    return (
      <Modal show
        containerClassName="modal-container"
        backdropClassName="modal-backdrop"
        onHide={this.props.onCancel}>
        <div className="modal-dialog">
          <div className="modal-header">
            <button className="modal-close" onClick={this.props.onCancel}>×</button>
            {this.state.title}
          </div>
          <div className="modal-body">
            {this.state.body}
          </div>
          <div className="modal-footer">
            <button className="js-cancel" onClick={this.props.onCancel}>{this.state.cancel}</button>
            <button className="button-secondary js-confirm" onClick={this.props.onConfirm}>{this.state.confirm}</button>
          </div>
        </div>
      </Modal>
    );
  }
}

ApiButtonModal.propTypes = {
  title: PropTypes.node,
  body: PropTypes.node,
  cancel: PropTypes.node,
  confirm: PropTypes.node,
  onCancel: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired,
};

export default ApiButton;
