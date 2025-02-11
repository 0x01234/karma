import React, { Component } from "react";
import PropTypes from "prop-types";

import { observable, action } from "mobx";
import { observer } from "mobx-react";

import semver from "semver";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons/faTrash";
import { faExclamationCircle } from "@fortawesome/free-solid-svg-icons/faExclamationCircle";
import { faCheckCircle } from "@fortawesome/free-solid-svg-icons/faCheckCircle";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons/faCircleNotch";

import { APIAlertmanagerUpstream } from "Models/API";
import { AlertStore, FormatBackendURI, FormatAlertsQ } from "Stores/AlertStore";
import { FormatQuery, QueryOperators, StaticLabels } from "Common/Query";
import { FetchWithCredentials } from "Common/Fetch";
import { Modal } from "Components/Modal";
import {
  LabelSetList,
  GroupListToUniqueLabelsList
} from "Components/LabelSetList";

const ProgressMessage = () => (
  <div className="text-center">
    <FontAwesomeIcon
      icon={faCircleNotch}
      className="text-muted display-1 mb-3"
      spin
    />
  </div>
);

const ErrorMessage = ({ message }) => (
  <div className="text-center">
    <FontAwesomeIcon
      icon={faExclamationCircle}
      className="text-danger display-1 mb-3"
    />
    <p>{message}</p>
  </div>
);
ErrorMessage.propTypes = {
  message: PropTypes.node.isRequired
};

const SuccessMessage = () => (
  <div className="text-center">
    <FontAwesomeIcon
      icon={faCheckCircle}
      className="text-success display-1 mb-3"
    />
    <p>
      Silence deleted, it might take a few minutes for affected alerts to change
      state
    </p>
  </div>
);

const DeleteSilenceModalContent = observer(
  class DeleteSilenceModalContent extends Component {
    static propTypes = {
      alertStore: PropTypes.instanceOf(AlertStore).isRequired,
      alertmanager: APIAlertmanagerUpstream.isRequired,
      silenceID: PropTypes.string.isRequired,
      onHide: PropTypes.func.isRequired
    };

    previewState = observable(
      {
        fetch: null,
        error: null,
        alertLabels: [],
        setError(err) {
          this.error = err;
        },
        groupsToUniqueLabels(groups) {
          this.alertLabels = GroupListToUniqueLabelsList(groups);
        }
      },
      {
        setError: action.bound,
        groupsToUniqueLabels: action.bound
      }
    );

    deleteState = observable(
      {
        fetch: null,
        done: false,
        error: null,
        setDone() {
          this.done = true;
        },
        setError(err) {
          this.error = err;
        },
        reset() {
          this.done = false;
          this.error = null;
        }
      },
      {
        setDone: action.bound,
        setError: action.bound,
        reset: action.bound
      }
    );

    parseAlertmanagerResponse = response => {
      /*
      {"status": "success"}
      or
      {
        "status": "error",
        "errorType": "bad_data",
        "error": "silence 706959fd-4590-4e21-b983-859ba6ec0e1a already expired"
      }
      */
      if (response.status === "success") {
        this.deleteState.setError(null);
      } else if (response.status === "error" && response.error) {
        this.deleteState.setError(response.error);
      } else {
        this.deleteState.setError(JSON.stringify(response));
      }
      this.deleteState.setDone();
    };

    onFetchPreview = () => {
      const { silenceID } = this.props;

      const alertsURI =
        FormatBackendURI("alerts.json?") +
        FormatAlertsQ([
          FormatQuery(StaticLabels.SilenceID, QueryOperators.Equal, silenceID)
        ]);

      this.previewState.fetch = FetchWithCredentials(alertsURI, {})
        .then(result => result.json())
        .then(result => {
          this.previewState.groupsToUniqueLabels(Object.values(result.groups));
          this.previewState.setError(null);
        })
        .catch(err => {
          console.trace(err);
          return this.previewState.setError(
            `Request fetching affected alerts failed with: ${err.message}`
          );
        });
    };

    onDelete = () => {
      const { alertmanager, silenceID } = this.props;

      // if it's already deleted then do nothing
      if (this.deleteState.done && this.deleteState.error === null) return;

      // reset state so we get a spinner
      this.deleteState.reset();

      const isOpenAPI = semver.satisfies(alertmanager.version, ">=0.16.0");

      const uri = isOpenAPI
        ? `${alertmanager.uri}/api/v2/silence/${silenceID}`
        : `${alertmanager.uri}/api/v1/silence/${silenceID}`;

      this.deleteState.fetch = FetchWithCredentials(uri, {
        method: "DELETE",
        headers: alertmanager.headers
      })
        .then(result => {
          if (isOpenAPI) {
            if (result.ok) {
              this.deleteState.setError(null);
              this.deleteState.setDone();
            } else {
              result.text().then(this.deleteState.setError);
              this.deleteState.setDone();
            }
          } else {
            result.json().then(this.parseAlertmanagerResponse);
          }
        })
        .catch(err => {
          console.trace(err);
          this.deleteState.setDone();
          this.deleteState.setError(
            `Delete request failed with: ${err.message}`
          );
        });
    };

    componentDidMount() {
      this.onFetchPreview();
    }

    render() {
      const { alertStore, onHide } = this.props;

      const isDone = this.deleteState.done && this.deleteState.error === null;

      return (
        <React.Fragment>
          <div className="modal-header">
            <h5 className="modal-title">Delete silence</h5>
            <button type="button" className="close" onClick={onHide}>
              <span>&times;</span>
            </button>
          </div>
          <div className="modal-body">
            {this.deleteState.done ? (
              this.deleteState.error !== null ? (
                <ErrorMessage message={this.deleteState.error} />
              ) : (
                <SuccessMessage />
              )
            ) : this.deleteState.fetch !== null ? (
              <ProgressMessage />
            ) : this.previewState.error === null ? (
              <LabelSetList
                alertStore={alertStore}
                labelsList={this.previewState.alertLabels}
              />
            ) : (
              <ErrorMessage message={this.previewState.error} />
            )}
            {isDone ? null : (
              <div className="d-flex flex-row-reverse">
                <button
                  type="button"
                  className="btn btn-outline-danger mr-2"
                  onClick={this.onDelete}
                  disabled={
                    this.deleteState.fetch !== null &&
                    this.deleteState.done === false
                  }
                >
                  <FontAwesomeIcon icon={faCheckCircle} className="mr-1" />
                  {this.deleteState.fetch !== null &&
                  this.deleteState.error !== null
                    ? "Retry"
                    : "Confirm"}
                </button>
              </div>
            )}
          </div>
        </React.Fragment>
      );
    }
  }
);

const DeleteSilence = observer(
  class DeleteSilence extends Component {
    static propTypes = {
      alertStore: PropTypes.instanceOf(AlertStore).isRequired,
      alertmanager: APIAlertmanagerUpstream.isRequired,
      silenceID: PropTypes.string.isRequired
    };

    toggle = observable(
      {
        visible: false,
        toggle() {
          this.visible = !this.visible;
        }
      },
      { toggle: action.bound }
    );

    render() {
      const { alertStore, alertmanager, silenceID } = this.props;

      return (
        <React.Fragment>
          <span
            className={`badge badge-danger cursor-pointer components-label components-label-with-hover`}
            onClick={this.toggle.toggle}
          >
            <FontAwesomeIcon className="mr-1" icon={faTrash} />
            Delete
          </span>
          <Modal isOpen={this.toggle.visible} toggleOpen={this.toggle.toggle}>
            <DeleteSilenceModalContent
              alertStore={alertStore}
              alertmanager={alertmanager}
              silenceID={silenceID}
              onHide={this.toggle.toggle}
            />
          </Modal>
        </React.Fragment>
      );
    }
  }
);

export { DeleteSilence, DeleteSilenceModalContent };
