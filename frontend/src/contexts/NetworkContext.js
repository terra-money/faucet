import React from 'react';

const NetworkContext = React.createContext({
  network: process.env.REACT_APP_CHAIN_ID,
  setNetwork: network => {
    console.log(network);
  }
});

export default NetworkContext;
