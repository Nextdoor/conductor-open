import React from 'react';
import moment from 'moment';

import PropTypes from 'prop-types';
import ApiButton from 'components/ApiButton';
import Card from 'components/Card';
import TrainComponent from 'components/TrainComponent';
import {trainProps, requestProps} from 'types/proptypes';

const PhaseTypes = {
  Delivery: 0,
  Verification: 1,
  Deploy: 2
};

class Phases extends TrainComponent {
  constructor(props) {
    super(props);
    this.state = this.initialState();
  }

  initialState() {
    return {
      focusedPhase: null,
      hoveredPhase: null,
      trainId: null
    };
  }

  componentWillReceiveProps(nextProps) {
    const {train} = nextProps;

    if (train === null) {
      // Reset state if train is null.
      this.setState(this.initialState());
    } else if (this.state.trainId !== train.id) {
      // Reset state if the train changed.
      const newState = this.initialState();
      newState.trainId = train.id;
      this.setState(newState);
    }
  }

  render() {
    return (
      <Card className="phases-card" header={this.getHeader()}>
        {this.getComponent()}
      </Card>
    );
  }

  getHeader() {
    const {train} = this.props;
    let deliveryPhaseId = null;
    let verificationPhaseId = null;
    let deployPhaseId = null;

    let activePhase = 0;
    if (train) {
      activePhase = train.active_phase;
      let activePhases = train.active_phases
      if (activePhases) {
        deliveryPhaseId = activePhases.delivery.id
        verificationPhaseId = activePhases.verification.id
        deployPhaseId = activePhases.deploy.id
      }
    }

    const focusedPhase = this.state.focusedPhase !== null ? this.state.focusedPhase : activePhase;
    const hoveredPhase = this.state.hoveredPhase;

    return (
      <div className="phases-header-tabs">
        {this.phaseTab(
          PhaseTypes.Delivery,
          focusedPhase,
          hoveredPhase,
          deliveryPhaseId)}
        {this.phaseTab(
          PhaseTypes.Verification,
          focusedPhase,
          hoveredPhase,
          verificationPhaseId)}
        {this.phaseTab(
          PhaseTypes.Deploy,
          focusedPhase,
          hoveredPhase,
          deployPhaseId)}
      </div>
    );
  }

  // This is a separate function rather than a React component
  // because this returns two sibling elements and using
  // a wrapper element would complicate the style.
  phaseTab(phaseType, focusedPhase, hoveredPhase, phaseId) {
    const focused = phaseType === focusedPhase;
    const hovered = phaseType === hoveredPhase;
    const focusable = !focused;

    let onClick = () => {
    };

    let extraClasses = '';
    if (focused) {
      extraClasses += ' focused';
    }
    if (hovered) {
      extraClasses += ' hovered';
    }
    if (focusable) {
      extraClasses += ' focusable';
      onClick = () => this.setState({focusedPhase: phaseType});
    }

    let text = '';
    switch (phaseType) {
      default:
      case PhaseTypes.Delivery:
        text = 'Delivery';
        break;
      case PhaseTypes.Verification:
        text = 'Verification';
        break;
      case PhaseTypes.Deploy:
        text = 'Deployment';
        break;
    }

    const onMouseEnter = () => this.setState({hoveredPhase: phaseType});
    const onMouseLeave = () => this.setState({hoveredPhase: null});

    return [
      <div className={'phases-header-tab' + extraClasses}
        key="tab"
        onClick={onClick}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}>
         <br /><span>{text}</span><br /><span className="job-id">ID: {phaseId}</span>
      </div>,
      <div className="phases-header-arrow-container"
        key="arrow"
        onClick={onClick}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}>
        <div className={'phases-header-arrow' + extraClasses}/>
      </div>
    ];
  }

  restartDeployButton() {
    // Only show for admins
    // If the deployment phase is the active phase
    // And deployment hasn't completed succesfully
    // as per backend API, phase can only be restarted on the latest train or the previous train.
    if (this.props.self.details.is_admin && this.props.train.active_phase === PhaseTypes.Deploy && !this.props.train.active_phases.deploy.completed_at) {
      return (
        <ApiButton
          modalProps={{
            title: 'Manually trigger deployment job',
            body: (
              <div>
                Would you like to manually kick off a deployment job?
                <br/><br/>
                Do this if the automatic deployment is not being triggered correctly.
              </div>
            )
          }}
          onClick={() => this.props.restartJob(this.props.train.id, 'deploy')}
          request={this.props.request}
          className="button-wide redploy-button">
          Restart Deploy
        </ApiButton>
      );
    }
  }

  getComponent() {
    const requestComponent = this.getRequestComponent();
    if (requestComponent !== null) {
      return requestComponent;
    }

    const {train} = this.props;

    const activePhase = this.state.focusedPhase !== null ? this.state.focusedPhase : train.active_phase;

    let jobs = [];
    let restart_deploy_componenet = null;
    switch (activePhase) {
      default:
      case PhaseTypes.Delivery:
        jobs = train.active_phases.delivery.jobs;
        break;
      case PhaseTypes.Verification:
        jobs = train.active_phases.verification.jobs;
        break;
      case PhaseTypes.Deploy:
        jobs = train.active_phases.deploy.jobs;
        restart_deploy_componenet = this.restartDeployButton()
        break;
    }

    const jobInfoList = [];
    for (let i = 0; i < jobs.length; i++) {
      const job = jobs[i];
      const jobInfo = {
        name: job.name.charAt(0).toUpperCase() + job.name.slice(1),
        id: job.id,
        url: job.url
      };

      if (job.completed_at !== null) {
        jobInfo.status = 'Completed in ' + moment(moment(job.completed_at) - moment(job.started_at)).format('mm:ss');
      } else if (job.started_at !== null) {
        jobInfo.status = 'Running for ' + moment(moment() - moment(job.started_at)).format('mm:ss');
      } else {
        jobInfo.status = 'Not yet started';
      }

      if (job.completed_at !== null) {
        switch (job.result) {
          case 0:
            jobInfo.result = <img src="/images/pass.png"/>;
            break;
          default:
          case 1:
            jobInfo.result = <img src="/images/fail.png"/>;
            break;
        }
      }

      jobInfoList.push(jobInfo);
    }

    const listItems = jobInfoList.map((job, i) => {
      const jobAttributes = (
        <div>
          <span className="job-name">{job.name}</span>
          <span className="job-result">{job.result}</span>
          <span className="job-status">{job.status}</span>
          <span className="job-id">ID: {job.id}</span>
        </div>
      );

      if (job.url !== null) {
        return (
          <li className="jobs-list-item link" key={i}>
            <a href={job.url}>
              <img className="job-link" src="/images/link.png"/>
              {jobAttributes}
            </a>
          </li>
        );
      }

      return (
        <li className="jobs-list-item" key={i}>
          {jobAttributes}
        </li>
      );
    });

    return (
      <div>
        <ul className="jobs-list">
          {listItems}
        </ul>
        <p className="phase-id">
          {restart_deploy_componenet}
        </p>
      </div>
    );
  }
}

Phases.propTypes = {train: trainProps,
  request: requestProps.isRequired,
  restartJob: PropTypes.func.isRequired,
};

export default Phases;
