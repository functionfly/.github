import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "./utils";
export function Slider({ value = [0], onValueChange, min = 0, max = 100, step = 1, disabled, className, }) {
    const [internalValue, setInternalValue] = React.useState(value);
    const currentValue = value ?? internalValue;
    const handleChange = (e) => {
        const newValue = [parseFloat(e.target.value)];
        setInternalValue(newValue);
        onValueChange?.(newValue);
    };
    const percentage = ((currentValue[0] - min) / (max - min)) * 100;
    return (_jsx("div", { className: cn("relative w-full", className), children: _jsx("input", { type: "range", min: min, max: max, step: step, value: currentValue[0], onChange: handleChange, disabled: disabled, className: "w-full h-2 bg-bg-tertiary rounded-full appearance-none cursor-pointer disabled:cursor-not-allowed disabled:opacity-50 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-brand-500 [&::-webkit-slider-thumb]:cursor-pointer", style: {
                background: `linear-gradient(to right, rgb(59, 130, 246) 0%, rgb(59, 130, 246) ${percentage}%, rgb(28, 28, 28) ${percentage}%, rgb(28, 28, 28) 100%)`,
            } }) }));
}
//# sourceMappingURL=Slider.js.map