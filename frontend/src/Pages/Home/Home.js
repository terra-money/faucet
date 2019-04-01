import React from 'react';
// import cx from "classnames";

import '../../App.scss';

function HomeComponent() {
  return (
    <div>
      <section>
        <h2>Terra Testnet Faucet</h2>
        <article>
          Hello intrepid spaceperson! Use this faucet to get tokens for the
          latest Terra testnet. Please don't abuse this service—the number of
          available tokens is limited.
        </article>
        <div className="recaptcha">
          <div>
            <div>
              <iframe
                title="recaptcha"
                src="https://www.google.com/recaptcha/api2/anchor?ar=2&amp;k=6Ld4w4cUAAAAAJceMYGpOTpjiJtMS_xvzOg643ix&amp;co=aHR0cDovL2ZhdWNldC50ZXJyYS5tb25leTo4MA..&amp;hl=ko&amp;v=v1552285980763&amp;size=normal&amp;cb=oiiw6vnnpfvz"
                width="304"
                height="78"
                role="presentation"
                name="a-hl7t8lstwbxg"
                frameborder="0"
                scrolling="no"
                sandbox="allow-forms allow-popups allow-same-origin allow-scripts allow-top-navigation allow-modals allow-popups-to-escape-sandbox"
              />
            </div>
          </div>
        </div>
        <div className="inputContainer">
          <div className="input">
            <input type="text" placeholder="Testnet address" />
          </div>
          <div className="select">
            <select>
              <option
                value=""
                disabled="disabled"
                selected="selected"
                hidden="hidden"
              >
                Select denom to receive...
              </option>
              <option value="luna">Luna</option>
              <option value="krw">KRW</option>
              <option value="usd">USD</option>
              <option value="sdr">SDR</option>
              <option value="gbp">GBP</option>
              <option value="eur">EUR</option>
              <option value="jpy">JPY</option>
              <option value="cny">CNY</option>
            </select>
            <div class="selectAddon">
              <i class="material-icons">arrow_drop_down</i>
            </div>
          </div>
        </div>
        <div className="buttonContainer">
          <button>
            <i aria-hidden="true" class="material-icons">
              send
            </i>
            <span>Send me tokens</span>
          </button>
        </div>
      </section>
      <section>
        <h2>Don't you have a testnet address?</h2>
        <article>
          There's two ways to get one. The first is by using Station, the crypto
          wallet for Terra. If you know command-line-fu, you can also generate
          an address with the Terra SDK.
        </article>
        <div className="buttonContainer">
          <button className="light">
            <i aria-hidden="true" class="material-icons">
              supervisor_account
            </i>
            <span>Join the latest testnet</span>
          </button>
        </div>
      </section>
    </div>
  );
}

export default HomeComponent;
