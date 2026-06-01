/**
 * CruftPromo — 30s (900 frame @ 30fps) 3-scene promo at 1920x1080.
 *
 *   A (0-270)   Hook    : scattered cache paths pile up behind the cruft wordmark
 *   B (270-660) Show     : a terminal scans in parallel; review list with risk chips
 *   C (660-900) Convert  : GB-reclaimed counter + --safe undo callout + install cmd
 *
 * Terminal / developer aesthetic. Monospace throughout; Inter for the tagline.
 */

import React from "react";
import {
  AbsoluteFill,
  Sequence,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";
import { loadFont as loadInter } from "@remotion/google-fonts/Inter";

const { fontFamily: inter } = loadInter();
const MONO = `'JetBrains Mono', 'SF Mono', 'Fira Code', ui-monospace, Menlo, monospace`;

const C = {
  ink: "#0b0e14",
  panel: "#0f1620",
  panelBar: "#161d28",
  green: "#7ee787",
  amber: "#f2cc60",
  red: "#f47067",
  blue: "#58a6ff",
  cream: "#e6edf3",
  dim: "#6e7681",
};

// Soft, static background glow (never animated during a scene).
const BrandBackground: React.FC = () => (
  <AbsoluteFill
    style={{
      background: C.ink,
      backgroundImage: `
        radial-gradient(55% 45% at 22% 28%, rgba(126,231,135,0.12), transparent 60%),
        radial-gradient(50% 50% at 80% 75%, rgba(88,166,255,0.12), transparent 60%)
      `,
    }}
  />
);

const fadeIn = (frame: number, start: number, len = 15) =>
  interpolate(frame, [start, start + len], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

const fadeOut = (frame: number, start: number, len = 20) =>
  interpolate(frame, [start, start + len], [1, 0], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

// ---- Scene A: Hook ---------------------------------------------------------

const SCATTERED = [
  { t: "~/Library/Developer/Xcode/DerivedData", x: 8, y: 16 },
  { t: "~/.ollama/models", x: 64, y: 12 },
  { t: "node_modules/", x: 30, y: 30 },
  { t: "~/.gradle/caches", x: 72, y: 34 },
  { t: "~/.cargo/registry", x: 12, y: 52 },
  { t: "~/.cache/huggingface/hub", x: 58, y: 56 },
  { t: "~/Library/Caches/pip", x: 80, y: 64 },
  { t: "~/.npm/_cacache", x: 18, y: 72 },
  { t: "target/", x: 44, y: 78 },
  { t: "~/.docker (dangling)", x: 70, y: 84 },
  { t: "~/Library/Caches/CocoaPods", x: 26, y: 88 },
  { t: "~/.bun/install/cache", x: 6, y: 36 },
];

const SceneA: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const markScale = spring({ frame, fps, config: { damping: 14 } });
  const cursorOn = Math.floor(frame / 15) % 2 === 0;
  const exit = fadeOut(frame, 245, 25);

  return (
    <AbsoluteFill style={{ opacity: exit }}>
      {/* scattered cruft */}
      {SCATTERED.map((p, i) => {
        const o = fadeIn(frame, 8 + i * 6, 18) * 0.5;
        const drift = interpolate(frame, [0, 270], [12, -12]);
        return (
          <div
            key={p.t}
            style={{
              position: "absolute",
              left: `${p.x}%`,
              top: `${p.y}%`,
              transform: `translateY(${drift}px)`,
              fontFamily: MONO,
              fontSize: 30,
              color: C.dim,
              opacity: o,
              whiteSpace: "nowrap",
            }}
          >
            {p.t}
          </div>
        );
      })}

      <AbsoluteFill style={{ justifyContent: "center", alignItems: "center" }}>
        <div
          style={{
            transform: `scale(${0.9 + markScale * 0.1})`,
            fontFamily: MONO,
            fontWeight: 800,
            fontSize: 180,
            letterSpacing: "-0.04em",
            color: C.cream,
          }}
        >
          <span style={{ color: C.green }}>$ </span>cruft
          <span
            style={{
              display: "inline-block",
              width: 90,
              height: 14,
              marginLeft: 18,
              background: cursorOn ? C.green : "transparent",
              verticalAlign: "middle",
            }}
          />
        </div>
        <div
          style={{
            marginTop: 28,
            fontFamily: inter,
            fontWeight: 600,
            fontSize: 56,
            color: C.cream,
            opacity: fadeIn(frame, 45, 25),
            transform: `translateY(${interpolate(frame, [45, 70], [24, 0], {
              extrapolateRight: "clamp",
            })}px)`,
            textAlign: "center",
          }}
        >
          Your Mac is hiding{" "}
          <span style={{ color: C.green }}>100+ GB</span> of dev cruft.
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

// ---- Scene B: Show ---------------------------------------------------------

type Row = { name: string; size: string; risk: "safe" | "risky" };
const ROWS: Row[] = [
  { name: "xcode-derived", size: "18.4 GB", risk: "safe" },
  { name: "ollama", size: "24.1 GB", risk: "risky" },
  { name: "project-artifacts  ×42", size: "12.7 GB", risk: "safe" },
  { name: "gradle", size: "6.2 GB", risk: "safe" },
  { name: "docker (prune)", size: "9.8 GB", risk: "safe" },
  { name: "huggingface", size: "8.0 GB", risk: "risky" },
];

const Chip: React.FC<{ risk: "safe" | "risky" }> = ({ risk }) => {
  const color = risk === "safe" ? C.green : C.amber;
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 10,
        fontFamily: MONO,
        fontSize: 26,
        color,
        border: `1px solid ${color}55`,
        background: `${color}14`,
        borderRadius: 999,
        padding: "4px 16px",
      }}
    >
      <span style={{ fontSize: 18 }}>●</span>
      {risk}
    </span>
  );
};

const SceneB: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const winScale = spring({ frame, fps, config: { damping: 18 } });
  const winOpacity = fadeIn(frame, 0, 18);
  const captionOpacity = fadeIn(frame, 10, 20);
  const exit = fadeOut(frame, 360, 25);

  // Running total ticks up as rows land.
  const total = interpolate(frame, [70, 250], [0, 79.2], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const spin = ["|", "/", "-", "\\"][Math.floor(frame / 4) % 4];
  const scanning = frame < 70;

  return (
    <AbsoluteFill
      style={{
        justifyContent: "center",
        alignItems: "center",
        opacity: exit,
      }}
    >
      <div
        style={{
          fontFamily: inter,
          fontWeight: 600,
          fontSize: 40,
          color: C.cream,
          opacity: captionOpacity,
          marginBottom: 28,
          textAlign: "center",
        }}
      >
        Scans every cache in parallel.{" "}
        <span style={{ color: C.dim }}>You tick what to delete.</span>
      </div>

      <div
        style={{
          width: 1280,
          borderRadius: 18,
          overflow: "hidden",
          background: C.panel,
          border: "1px solid #233042",
          boxShadow: "0 50px 140px rgba(0,0,0,0.55)",
          opacity: winOpacity,
          transform: `scale(${0.96 + winScale * 0.04})`,
        }}
      >
        {/* window bar */}
        <div
          style={{
            height: 56,
            background: C.panelBar,
            display: "flex",
            alignItems: "center",
            padding: "0 22px",
            gap: 12,
          }}
        >
          {[C.red, C.amber, C.green].map((d) => (
            <div
              key={d}
              style={{ width: 16, height: 16, borderRadius: 999, background: d }}
            />
          ))}
          <div
            style={{
              fontFamily: MONO,
              fontSize: 24,
              color: C.dim,
              marginLeft: 16,
            }}
          >
            cruft — review
          </div>
        </div>

        {/* body */}
        <div style={{ padding: "34px 44px", fontFamily: MONO }}>
          <div style={{ fontSize: 30, color: C.cream, marginBottom: 10 }}>
            <span style={{ color: C.blue }}>$</span> cruft
          </div>
          <div style={{ fontSize: 28, color: C.dim, marginBottom: 26 }}>
            {scanning
              ? `${spin} scanning 57 cleaners in parallel…`
              : "✓ scan complete — 57 cleaners, 6 with reclaimable space"}
          </div>

          {ROWS.map((r, i) => {
            const appear = 60 + i * 22;
            const o = fadeIn(frame, appear, 12);
            const x = interpolate(frame, [appear, appear + 12], [-24, 0], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
            });
            return (
              <div
                key={r.name}
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  padding: "14px 0",
                  borderBottom: "1px solid #1b2533",
                  opacity: o,
                  transform: `translateX(${x}px)`,
                }}
              >
                <span style={{ fontSize: 30, color: C.cream }}>
                  <span style={{ color: C.green, marginRight: 16 }}>[x]</span>
                  {r.name}
                </span>
                <span style={{ display: "flex", alignItems: "center", gap: 28 }}>
                  <span style={{ fontSize: 30, color: C.cream, width: 150, textAlign: "right" }}>
                    {r.size}
                  </span>
                  <Chip risk={r.risk} />
                </span>
              </div>
            );
          })}

          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              marginTop: 28,
              fontSize: 34,
              color: C.cream,
            }}
          >
            <span style={{ color: C.dim }}>reclaimable</span>
            <span style={{ color: C.green, fontWeight: 700 }}>
              {total.toFixed(1)} GB
            </span>
          </div>
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ---- Scene C: Convert ------------------------------------------------------

const SceneC: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const count = interpolate(frame, [0, 70], [0, 84.2], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const numScale = spring({ frame, fps, config: { damping: 13 } });
  const subOpacity = fadeIn(frame, 55, 25);
  const cmdOpacity = fadeIn(frame, 95, 25);

  return (
    <AbsoluteFill style={{ justifyContent: "center", alignItems: "center" }}>
      <div
        style={{
          fontFamily: MONO,
          fontWeight: 800,
          fontSize: 220,
          color: C.green,
          letterSpacing: "-0.04em",
          transform: `scale(${0.9 + numScale * 0.1})`,
          textShadow: "0 0 80px rgba(126,231,135,0.35)",
        }}
      >
        {count.toFixed(1)} GB
      </div>
      <div
        style={{
          fontFamily: inter,
          fontWeight: 600,
          fontSize: 52,
          color: C.cream,
          marginTop: 8,
        }}
      >
        reclaimed.
      </div>
      <div
        style={{
          fontFamily: inter,
          fontSize: 38,
          color: C.dim,
          marginTop: 28,
          opacity: subOpacity,
          textAlign: "center",
        }}
      >
        Deleted on confirm ·{" "}
        <span style={{ color: C.amber }}>--safe</span> keeps a 7-day undo.
      </div>

      <div
        style={{
          marginTop: 56,
          opacity: cmdOpacity,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 20,
        }}
      >
        <div
          style={{
            fontFamily: MONO,
            fontSize: 40,
            color: C.cream,
            background: C.panel,
            border: "1px solid #233042",
            borderRadius: 14,
            padding: "18px 34px",
          }}
        >
          <span style={{ color: C.green }}>$ </span>
          brew install sachincool/tap/cruft
        </div>
        <div style={{ fontFamily: inter, fontSize: 34, color: C.blue }}>
          github.com/sachincool/cruft
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ---- Composition -----------------------------------------------------------

export const CruftPromo: React.FC = () => {
  return (
    <AbsoluteFill style={{ background: C.ink }}>
      <BrandBackground />
      <Sequence from={0} durationInFrames={270} name="A: Hook">
        <SceneA />
      </Sequence>
      <Sequence from={270} durationInFrames={390} name="B: Show">
        <SceneB />
      </Sequence>
      <Sequence from={660} durationInFrames={240} name="C: Convert">
        <SceneC />
      </Sequence>
    </AbsoluteFill>
  );
};
