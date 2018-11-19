import React from 'react';
import PropTypes from 'prop-types';

import ApiButton from 'components/ApiButton';
import Card from 'components/Card';
import Error from 'components/Error';
import Loading from 'components/Loading';
import {configProps, trainProps, requestProps} from 'types/proptypes';
import {Modes} from 'types/config';

class Admin extends React.Component {
  componentWillMount() {
    const {fetchConfig, fetchConfigRequest} = this.props;

    if (fetchConfigRequest.fetching !== true && fetchConfigRequest.receivedAt === null) {
      fetchConfig();
    }
  }

  render() {
    const {self, fetchConfigRequest} = this.props;

    if (fetchConfigRequest.fetching !== true && fetchConfigRequest.receivedAt === null) {
      return null;
    }

    if (self === null || self.is_admin === false) {
      return null;
    }

    return (
      <Card header="Admin Tools">
        {this.getComponent()}
      </Card>
    );
  }

  getComponent() {
    const {train, config, fetchConfigRequest} = this.props;

    if (fetchConfigRequest.error !== null) {
      return <Error message={fetchConfigRequest.error}/>;
    }
    if (train === null || config === null) {
      return <Loading/>;
    }

    return (
      <div className="admin-tools">
        {this.toggleModeButton()}
        {this.toggleCloseButton()}
      </div>
    );
  }

  toggleModeButton() {
    let title = '';
    let body = '';
    if (this.props.config.mode === Modes.Schedule) {
      title = 'Confirm switch to manual mode';
      body = (
        <div>
          Manual mode disables automatically closing and opening trains based on schedule.
          <br/><br/>
          There should be a good reason to switch to manual mode, such as a large infrastructure outage,
          and it should be switched back when the situation is resolved.
        </div>
      );
    } else {
      title = 'Confirm switch to schedule mode';
      body = (
        <div>
          Schedule mode automatically closes and opens trains based on time of day.
          <br/><br/>
          This is the preferred mode if there are no circumstances necessitating manual mode.
        </div>
      );
    }
    return (
      <ApiButton
        modalProps={{
          title: title,
          body: body
        }}
        onClick={this.props.toggleMode}
        request={this.props.toggleModeRequest}
        className="admin-api-button button-wide js-toggle-mode">
        {this.props.config.mode === Modes.Schedule ? 'Manual mode' : 'Schedule mode'}
      </ApiButton>);
  }

  toggleCloseButton() {
    if (!this.isTrainCloseToggleable(this.props.train)) {
      return null;
    }
    let title = '';
    let body = '';

    if (this.props.train.closed) {
      title = 'Confirm train open';
      body = (
        <div>
          Opening the train will override automatic closing and
          result in the train continuously extending as changes come in.
          It will also prevent the train from being deployed.
          <br/><br/>
          There should be a good reason to open the train. Consider extending or using manual mode.
        </div>
      );
    } else {
      title = 'Confirm train close';
      body = (
        <div>
          Closing the train will override automatic closing and allow it to be deployed if verified.
          <br/><br/>
          There should be a good reason to close the train, such as deploying after hours or in manual mode.
        </div>
      );
    }
    return (
      <ApiButton
        modalProps={{
          title: title,
          body: body
        }}
        onClick={this.props.toggleClose}
        request={this.props.toggleCloseRequest}
        className="admin-api-button js-toggle-lock">
        {this.props.train.closed ? 'Open' : 'Close'}
      </ApiButton>
    );
  }

  isTrainCloseToggleable(train) {
    return !train.done && train.active_phases.deploy.started_at === null;
  }
}

Admin.propTypes = {
  self: PropTypes.shape({
    name: PropTypes.string.isRequired,
    email: PropTypes.string.isRequired,
    is_admin: PropTypes.bool.isRequired,
  }),
  train: trainProps,
  config: configProps,

  fetchConfig: PropTypes.func.isRequired,
  fetchConfigRequest: requestProps.isRequired,

  toggleMode: PropTypes.func.isRequired,
  toggleModeRequest: requestProps.isRequired,

  toggleClose: PropTypes.func.isRequired,
  toggleCloseRequest: requestProps.isRequired,
};

export default Admin;
