/* eslint-disable no-invalid-this */

import React from 'react';
import {mount} from 'enzyme';
import Summary from 'components/Summary';
import {newTrain, noRequest, completeRequest, user} from 'test/TestData';
import {Phases} from 'types/train';

describe('Summary', function() {

  beforeEach(function() {
    this.user = JSON.parse(JSON.stringify(user));
    this.train = JSON.parse(JSON.stringify(newTrain));
  });

  it('Delivering renders correctly', function() {
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        request={completeRequest}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('Delivering new changes to staging'));

    // Users with changes
    expect(wrapper.text()).toEqual(expect.stringContaining('Bob Bailey'));
    expect(wrapper.text()).toEqual(expect.stringContaining('Joe Smith'));
  });

  it('Extending renders correctly', function() {
    this.train.all_phase_groups = [this.train.all_phase_groups[0], this.train.all_phase_groups[0]];
    this.train.all_phase_groups[0].delivery.completed_at = '2000-01-01T00:00:00Z';
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        request={completeRequest}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('Train extended'));
  });

  it('Verifying renders correctly', function() {
    this.train.active_phase = Phases.Verification;
    this.train.not_deployable_reason = 'test reason';
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        request={completeRequest}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('test reason'));
  });

  it('Deploying renders correctly', function() {
    this.train.active_phase = Phases.Deploy;
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        request={completeRequest}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('Deploying to production'));
  });

  it('Deployed renders correctly', function() {
    this.train.done = true;
    this.train.active_phase = Phases.Deploy;
    this.train.deployed_at = '2000-01-01T00:00:00Z';
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        request={completeRequest}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('deployed to production'));
  });

  it('Cancelled renders correctly', function() {
    this.train.done = true;
    this.train.active_phase = Phases.Deploy;
    this.train.deployed_at = null;
    this.train.cancelled_at = '2000-01-01T00:00:00Z';
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        request={completeRequest}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('cancelled on'));
  });

  it('Never finished renders correctly', function() {
    this.train.done = true;
    this.train.active_phase = Phases.Deploy;
    this.train.deployed_at = null;
    this.train.cancelled_at = null;
    this.train.active_phases.deploy.started_at = null;
    const wrapper = mount(
      <Summary
        self={user}
        train={this.train}
        request={completeRequest}
        extendTrain={() => {}}
        blockTrain={() => {}}
        unblockTrain={() => {}}
        cancelTrain={() => {}}
        rollbackToTrain={() => {}}
        requestExtend={noRequest}
        requestBlock={noRequest}
        requestUnblock={noRequest}
        requestCancel={noRequest}
        requestRollback={noRequest}/>);
    expect(wrapper.text()).toEqual(expect.stringContaining('never deployed'));
  });

});
