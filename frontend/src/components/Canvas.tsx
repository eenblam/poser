import { useEffect, useRef } from 'react';

/*
// https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons
enum MouseButtons {
    None = 0,
    Primary = 1,
    Secondary = 2,
    Auxiliary = 4,
    Back = 8,
    Forward = 16,
}
*/

// https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/button
enum MouseButton {
    Primary = 0,
    Secondary = 1,
    Auxiliary = 2,
    Back = 3,
    Forward = 4,
}

function Canvas() {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        if (canvasRef.current === null) { return; }
        const canvas = canvasRef.current;

        const context = canvas.getContext('2d');
        if (context === null) { return; }

        // Get a non-null context to avoid constantly checking later.
        const ctx: CanvasRenderingContext2D = context;

        canvas.style.background = 'white';
        ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';

        /*
        function resize() {
            // Busted, clears image each time.
            // Using canvas.getImageData/putImageData doesn't re-scale.
            // Using an Image allows scaling, but it's lossy.
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
            ctx.clearRect(0,   0, canvas.width, canvas.height);
        }
        window.onresize = resize;
        resize()
        */
        //canvas.width = window.innerWidth;
        //canvas.height = window.innerHeight;
        ctx.clearRect(0,   0, canvas.width, canvas.height);

        let lastPoint = { x: 0, y: 0, first: true}
        let drawing = false;
        function move(e: MouseEvent) {
            if (drawing) {
                if (lastPoint.first) {
                    lastPoint = {x: e.offsetX, y: e.offsetY, first: false}
                }

                ctx.beginPath();
                ctx.moveTo(lastPoint.x, lastPoint.y);
                ctx.lineTo(e.offsetX, e.offsetY);

                ctx.strokeStyle = 'pink'; //TODO player color
                ctx.lineWidth = 5;
                ctx.lineCap = 'round';
                ctx.stroke();
                ctx.closePath();
                console.log(lastPoint);

                lastPoint = {x: e.offsetX, y: e.offsetY, first: false}
            }
        }
        canvas.onmousemove = move;
        canvas.onmousedown = (e: MouseEvent) => {
            if (e.button != MouseButton.Primary) { return; }
            lastPoint.first = true;
            drawing = true;
        }
        canvas.onmouseup = (e: MouseEvent) => {
            if (e.button != MouseButton.Primary) { return; }
            lastPoint.first = true;
            drawing = false;
        }
        canvas.onmouseleave = (_: MouseEvent) => {
            // If we've left the canvas, stop drawing.
            lastPoint.first = true;
            drawing = false;
        }

        return () => { // cleanup
            window.onresize = null;
        };

    }, []);

    return (
        <canvas ref={canvasRef}></canvas>
    )
}

export default Canvas
