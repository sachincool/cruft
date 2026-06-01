import React from "react";
import { Composition } from "remotion";
import { CruftPromo } from "./CruftPromo";

const WIDTH = 1920;
const HEIGHT = 1080;
const FPS = 30;
const DURATION_IN_FRAMES = FPS * 30; // 30s = 900 frames

export const RemotionRoot: React.FC = () => {
  return (
    <Composition
      id="CruftPromo"
      component={CruftPromo}
      durationInFrames={DURATION_IN_FRAMES}
      fps={FPS}
      width={WIDTH}
      height={HEIGHT}
    />
  );
};
