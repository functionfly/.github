import * as React from "react";
export interface SliderProps {
    value?: number[];
    onValueChange?: (value: number[]) => void;
    min?: number;
    max?: number;
    step?: number;
    disabled?: boolean;
    className?: string;
}
export declare function Slider({ value, onValueChange, min, max, step, disabled, className, }: SliderProps): React.JSX.Element;
//# sourceMappingURL=Slider.d.ts.map