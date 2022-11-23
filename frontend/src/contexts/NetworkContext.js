import React from 'react';
import { networks } from '../config';

const NetworkContext = React.createContext({
  network: networks[0].chainID,
  setNetwork: (network) => {
    console.log(network);
  },
});

export default NetworkContext;
