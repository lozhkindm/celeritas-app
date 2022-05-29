<?php

[$firstDayNumber, $lastDayNumber, $currentDayNumber, $recordsInDB] = explode(' ', readline());

$graphSize = $lastDayNumber - $firstDayNumber + 1;
$graph = array_fill($firstDayNumber, $graphSize, null);

for ($i = 1; $i <= $recordsInDB; $i++) {
    [$recordKey, $recordValue] = explode(' ', readline());

    $graph[$recordKey] = (int) $recordValue;
}

$lastValue = null;

foreach ($graph as $recordKey => $recordValue) {
    if ($recordValue === null || $recordValue === -1) {
        if ($lastValue === null || $recordKey > $currentDayNumber) {
            $graph[$recordKey] = -1;
        } else {
            $graph[$recordKey] = $lastValue;
        }
    } else {
        $lastValue = $recordValue;
    }
}

foreach ($graph as $recordKey => $recordValue) {
    echo sprintf('%d %d', $recordKey, $recordValue) . PHP_EOL;
}
