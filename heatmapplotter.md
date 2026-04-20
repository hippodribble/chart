# HeatMap

- a set of rectangles that are coloured and displayed in order on a grid with minimal spacing
- there is a fixed height to the plot, ie a fixed number of markers. These might represent hours of the day or days of the week

## Data Definition

- a set of values to be plotted as rectangles by colour
- the number of rectangles in each column
- colours representing min and max values, or 'auto'
- meaningful labels for the axes

```
[]string labels for vertical axis - the count determines the chart height in samples
[]string labels for horizontal axis - there must be exactly enough of them (crop or pad the list)
*[2]color.RGBA - RGBA for low and high colours - if nil, assign automatically to blue/red
float32 spacingX,spacingY - gaps between shapes, as a percentage of the data point size
float32 aspectRatio - width:height - in case squares aren't doing it
int pointsPerColumn - so the correct column can be used


```


