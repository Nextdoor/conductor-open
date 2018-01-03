/* eslint-disable no-invalid-this */
import React from 'react';
import {mount} from 'enzyme';
import {newTrain, noRequest, completeRequest, configSchedule, configManual, user, adminUser} from 'test/TestData';
import Admin from "./Admin";

describe('Admin', function() {
  beforeEach(function() {
    this.user = JSON.parse(JSON.stringify(user));
    this.adminUser = JSON.parse(JSON.stringify(adminUser));
    this.train = JSON.parse(JSON.stringify(newTrain));
    this.configSchedule = JSON.parse(JSON.stringify(configSchedule));
    this.configManual = JSON.parse(JSON.stringify(configManual));
  });

  it('Is inaccessible to non-admin users', function() {
    const wrapper = mount(
      <Admin
        self={this.user}
        train={this.train}
        config={this.configSchedule}
        fetchConfig={() => {}}
        fetchConfigRequest={completeRequest}
        toggleMode={() => {}}
        toggleModeRequest={noRequest}
        toggleClose={() => {}}
        toggleCloseRequest={noRequest} />);
    expect(wrapper.children().length).toEqual(0);
  });

  it('Waits for train gracefully', function() {
    const wrapper = mount(
      <Admin
        self={this.adminUser}
        train={null}
        config={this.configSchedule}
        fetchConfig={() => {}}
        fetchConfigRequest={completeRequest}
        toggleMode={() => {}}
        toggleModeRequest={noRequest}
        toggleClose={() => {}}
        toggleCloseRequest={noRequest} />);
    expect(wrapper.find('.loading').length).toEqual(1);
  });

  it('Waits for config gracefully', function(done) {
    const wrapper = mount(
      <Admin
        self={this.adminUser}
        train={this.train}
        config={null}
        fetchConfig={done}
        fetchConfigRequest={noRequest}
        toggleMode={() => {}}
        toggleModeRequest={noRequest}
        toggleClose={() => {}}
        toggleCloseRequest={noRequest} />);
    expect(wrapper.find('.loading').length).toEqual(1);
  });

  it('Renders correctly', function() {
    const wrapper = mount(
      <Admin
        self={this.adminUser}
        train={this.train}
        config={this.configSchedule}
        fetchConfig={() => {}}
        fetchConfigRequest={completeRequest}
        toggleMode={() => {}}
        toggleModeRequest={noRequest}
        toggleClose={() => {}}
        toggleCloseRequest={noRequest} />);
    expect(wrapper.text()).toEqual(expect.stringContaining('Admin Tools'));
    expect(wrapper.find('.loading').length).toEqual(0);
    expect(wrapper.text()).toEqual(expect.stringContaining('Manual mode'));
    expect(wrapper.text()).toEqual(expect.stringContaining('Close'));
  });

  it('Allows mode toggling', function() {
    const wrapper = mount(
      <Admin
        self={this.adminUser}
        train={this.train}
        config={this.configSchedule}
        fetchConfig={() => {}}
        fetchConfigRequest={completeRequest}
        toggleMode={() => {}}
        toggleModeRequest={noRequest}
        toggleClose={() => {}}
        toggleCloseRequest={noRequest} />);
    expect(wrapper.find('.js-toggle-mode').length).toEqual(1);
  });

  it('Allows train lock toggling', function() {
    const wrapper = mount(
      <Admin
        self={this.adminUser}
        train={this.train}
        config={this.configSchedule}
        fetchConfig={() => {}}
        fetchConfigRequest={completeRequest}
        toggleMode={() => {}}
        toggleModeRequest={noRequest}
        toggleClose={() => {}}
        toggleCloseRequest={noRequest} />);
    expect(wrapper.find('.js-toggle-lock').length).toEqual(1);
  });
});
