import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import './index.scss';
import App from './App';
import * as serviceWorker from './serviceWorker';
import NetworkContext from './contexts/NetworkContext';
import { networks } from './config';

class Root extends Component {
  setNetwork = (network) => {
    this.setState({
      network: network,
    });
  };
  state = {
    network: networks[0].key,
    setNetwork: this.setNetwork,
  };

  render() {
    return (
      <NetworkContext.Provider value={this.state}>
        <Route>
          <App />
        </Route>
      </NetworkContext.Provider>
    );
  }
}

ReactDOM.render(
  <Router>
    <Root />
  </Router>,
  document.getElementById('root')
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
