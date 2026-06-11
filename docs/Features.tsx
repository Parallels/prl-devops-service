import { useRef } from "react";
import { motion, useScroll, useTransform, useSpring, MotionValue } from "framer-motion";

interface FeatureSlide {
  id: number;
  pre: string;
  title: string;
  description: string;
  image: string;
  color: string;
  layout: "left" | "right";
}

const features: FeatureSlide[] = [
  {
    id: 1,
    pre: "Discover",
    title: "One‑Click Application Marketplace",
    description: "Browse, select, and launch OS-level isolated applications in seconds, with sensible defaults that just work.",
    image: "/assets/features/slide-1.png",
    color: "#3B82F6",
    layout: "right",
  },
  {
    id: 2,
    pre: "Built on",
    title: "True Isolation by Design",
    description: "Each capsule runs fully isolated using OS‑level virtualization, protecting your data, workloads, and network boundaries.",
    image: "/assets/features/slide-2.png",
    color: "#8B5CF6",
    layout: "left",
  },
  {
    id: 3,
    pre: "It just works!",
    title: "Zero Setup. Zero Disruption.",
    description: "Drop Capsules into your existing workflow, no rewiring, no new tools, no friction.",
    image: "/assets/features/slide-3.png",
    color: "#EC4899",
    layout: "right",
  },
  {
    id: 4,
    pre: "Lightweight",
    title: "Use Only What You Need",
    description: "Capsules start instantly, sleep when idle, and never waste resources.",
    image: "/assets/features/slide-4.png",
    color: "#10B981",
    layout: "left",
  },
  {
    id: 5,
    pre: "Sophisticated",
    title: "Built‑In Networking & Domains",
    description: "Every capsule gets predictable networking, automatic routing, and its own domain, no manual proxying required.",
    image: "/assets/features/slide-5.png",
    color: "#14B8A6",
    layout: "right",
  },
  {
    id: 6,
    pre: "Community",
    title: "Blueprints That Evolve",
    description: "Launch proven configurations built by the community, and share your own.",
    image: "/assets/features/slide-6.png",
    color: "#22C55E",
    layout: "left",
  },
  {
    id: 7,
    pre: "It's ready",
    title: "Built for What Comes Next",
    description: "Powered by open standards and modern runtimes, no lock‑in, no dead ends.",
    image: "/assets/features/slide-7.png",
    color: "#64748B",
    layout: "right",
  },
];

const featureShift = -0.2; // Shift earlier to reduce first slide hold time

// Sub-component for individual slides
const FeatureSlide = ({ feature, index, total, scrollYProgress }: { feature: FeatureSlide; index: number; total: number; scrollYProgress: MotionValue<number> }) => {
  // Distribute slides evenly from 0 (start) to 1 (end)
  const step = 1 / Math.max(total - 1, 1);
  const center = (index + featureShift) * step;

  const gap = step * 0.6;
  const start = center - gap;
  const end = center + gap;
  const fade = step * 0.2;

  const range = [start, start + fade, end - fade, end];
  const opacity = useTransform(scrollYProgress, range, [0, 1, 1, 0]);

  // OPTION B: PARALLAX MOVEMENTS
  // Text moves slow/up, Image moves fast/down or vice versa
  const yText = useTransform(scrollYProgress, range, [200, 0, 0, -200]); // Faster movement
  const yImage = useTransform(scrollYProgress, range, [100, 0, 0, -100]); // Slower, feels deeper

  const scale = useTransform(scrollYProgress, range, [0.85, 1, 1, 0.85]);

  const isLeft = feature.layout === "left";

  return (
    <motion.div className="absolute inset-0 w-full h-full flex items-center justify-center p-0 md:p-6" style={{ opacity, zIndex: index }}>
      {/* 
               OPTION D: GLASS PARALLAX
               Same parallax structure as Option 5, but the text container is wrapped 
               in a heavy glass card to ensure perfect readability over overlap.
            */}

      <div className="relative w-full max-w-7xl h-full md:h-[80vh] flex items-center justify-center">
        {/* Layer 1: The Image (Background, Large) */}
        <motion.div
          className={`absolute top-0 w-full h-[50vh] md:h-full md:w-[70%] z-0 rounded-b-3xl md:rounded-3xl overflow-hidden shadow-2xl opacity-100
                        left-0 right-0 md:bottom-0
                        ${isLeft ? "md:right-0 md:left-auto md:origin-right" : "md:left-0 md:right-auto md:origin-left"}
                    `}
          style={{
            y: yImage,
            scale,
            boxShadow: `0 25px 50px -12px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.1) inset, 0 0 30px -5px ${feature.color}40`,
          }}
        >
          <img src={feature.image} alt={feature.title} className="w-full h-full object-cover transition-transform duration-700 hover:scale-105" />
          {/* Subtle Dark Overlay for contrast */}
          <div className="absolute inset-0 bg-slate-900/10 mix-blend-multiply" />

          {/* Gloss/Reflection Overlay */}
          <div className="absolute inset-0 bg-gradient-to-tr from-transparent via-transparent to-white/30 opacity-100 pointer-events-none" />
        </motion.div>

        {/* Layer 2: The Text (Foreground, Floating, GLASS) */}
        <motion.div
          className={`absolute z-10 w-auto max-w-[90%] md:max-w-[55%] flex flex-col justify-center
                        ${isLeft ? "left-4 md:left-12 text-left" : "right-4 md:right-12 text-right"}
                    `}
          style={{ y: yText }}
        >
          {/* GLASS Text Block - Premium Slate Gradient */}
          <div
            className="relative overflow-hidden p-10 md:p-14 rounded-3xl backdrop-blur-2xl border border-white/30 ring-1 ring-white/50 transition-shadow duration-500"
            style={{
              background: "linear-gradient(145deg, rgba(15, 23, 42, 0.05) 0%, rgba(255, 255, 255, 0.8) 40%, rgba(255, 255, 255, 0.6) 100%)",
              boxShadow: `0 30px 60px -15px rgba(0,0,0,0.2), 0 0 0 1px rgba(255,255,255,0.4) inset, 0 20px 40px -10px ${feature.color}30`,
            }}
          >
            {/* Dither Noise Overlay */}
            <div
              className="absolute inset-0 opacity-[0.4] pointer-events-none mix-blend-overlay"
              style={{
                backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E")`,
              }}
            />

            <div
              className={`relative z-10 flex items-center gap-3 mb-6 opacity-100
                             ${isLeft ? "justify-start" : "justify-end"}
                        `}
            >
              <span className="text-sm font-mono tracking-[0.3em] uppercase text-blue-600 font-extrabold  px-3 py-1 rounded">{feature.pre}</span>
            </div>
            <h1 className="relative z-10 text-5xl md:text-6xl lg:text-7xl font-black tracking-tighter leading-none text-transparent bg-clip-text bg-gradient-to-r from-slate-900 to-slate-600 drop-shadow-sm mb-6 pb-3 pr-2">
              {feature.title}
            </h1>
            <p className="relative z-10 text-lg md:text-xl text-slate-700 font-medium leading-relaxed">{feature.description}</p>
          </div>
        </motion.div>
      </div>
    </motion.div>
  );
};

// Controls how much scroll distance (in vh) is needed per slide.
// Lower = faster transitions (less scroll needed).
// Higher = slower transitions (more scroll needed).
// Range: 10 (very fast) to 100+ (very slow).
const SCROLL_SENSITIVITY = 15;
const SHOW_SLIDE_INDICATORS = true;

export const Features = () => {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });

  // SMOOTH SCROLLING PHYSICS
  // Wraps the raw scroll value with a spring physics simulation.
  // This decouples the animation from the raw mouse wheel input, removing "jumps".
  const smoothProgress = useSpring(scrollYProgress, {
    mass: 0.1,
    stiffness: 100,
    damping: 20,
    restDelta: 0.001,
  });

  const colorInput = features.map((_, i) => (i + featureShift) / (features.length - 1));
  const colorOutput = features.map((f) => f.color);
  const backgroundColor = useTransform(smoothProgress, colorInput, colorOutput);

  // Calculate total height: One viewport for the sticky content + scrollable distance per feature
  const totalHeight = 100 + features.length * SCROLL_SENSITIVITY;

  return (
    <div ref={containerRef} style={{ height: `${totalHeight}vh` }} className="relative">
      <div className="sticky top-16 h-[calc(100vh-4rem)] overflow-hidden">
        <div className="absolute inset-0 bg-slate-50" />

        <motion.div
          className="absolute inset-0 transition-colors duration-700 ease-linear"
          style={{
            background: useTransform(backgroundColor, (color) => `radial-gradient(circle at 50% 50%, ${color}10 0%, transparent 50%)`),
            willChange: "background", // Performance hint
          }}
        />

        <div
          className="absolute inset-0 opacity-[0.12] pointer-events-none"
          style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 1000 1000' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E")`,
          }}
        />

        <div className="relative w-full h-full">
          {features.map((feature, i) => (
            <FeatureSlide key={feature.id} feature={feature} index={i} total={features.length} scrollYProgress={smoothProgress} />
          ))}
        </div>

        <motion.div className="absolute bottom-0 left-0 h-1.5 bg-gradient-to-r from-cyan-400 via-blue-500 to-purple-600 z-50 origin-left" style={{ scaleX: smoothProgress }} />

        {/* Floating Slide Position Indicator */}
        {SHOW_SLIDE_INDICATORS && (
          <div className="absolute bottom-8 left-1/2 -translate-x-1/2 flex items-center gap-3 z-50 bg-white/20 backdrop-blur-md px-4 py-2 rounded-full border border-white/10 shadow-lg">
            {features.map((_, i) => (
              <SlideIndicatorDot key={i} index={i} total={features.length} scrollYProgress={smoothProgress} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

const SlideIndicatorDot = ({ index, total, scrollYProgress }: { index: number; total: number; scrollYProgress: MotionValue<number> }) => {
  const step = 1 / Math.max(total - 1, 1);
  // Use the same shift logic as the main slides to match active state
  const center = (index + featureShift) * step;

  // Narrower range for sharper activation
  const range = [center - step * 0.5, center, center + step * 0.5];

  const opacity = useTransform(scrollYProgress, range, [0.3, 1, 0.3]);
  const scale = useTransform(scrollYProgress, range, [0.8, 1.2, 0.8]);
  const background = useTransform(scrollYProgress, range, ["#94a3b8", "#3b82f6", "#94a3b8"]); // Slate-400 to Blue-500

  return (
    <motion.div
      className="w-2.5 h-2.5 rounded-full cursor-pointer"
      style={{
        opacity,
        scale,
        backgroundColor: background,
      }}
    />
  );
};
