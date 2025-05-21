import React from 'react';
import { Provider, defaultTheme, Button } from '@adobe/react-spectrum';

function App() {
  return (
    <Provider theme={defaultTheme}>
      <Button variant="cta" onPress={() => alert('Button pressed!')}>
        Hello React Spectrum!
      </Button>
    </Provider>
  );
}

export default App;
