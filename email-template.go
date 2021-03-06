package main

const EMAIL_TEMPLATE = `
<!DOCTYPE html>
<html>

<head>
    <style>
        table {
            font-family: arial, sans-serif;
            border-collapse: collapse;
            width: 100%;
        }

        td,
        th {
            border: 1px solid #dddddd;
            text-align: left;
            padding: 8px;
        }

        tr:nth-child(even) {
            background-color: #dddddd;
        }
    </style>
</head>

<body>
    <h2>Vaccine Availability</h2>
    <h4>The slots get booked very fast. Please hurry booking process for better possibility</h4>
    <table>
        <tr>
            <th>Center Name</th>
            <th>Availabe Slots</th>
            <th>Date</th>
            <th>Vaccine Type</th>
            <th>Age</th>
            <th>Fee Type</th>
        </tr>
        {{range $index, $slot := .}}
        <tr>
            <td>{{ $slot.Center }}</td>
            <td>{{ $slot.AvailableSlots }}</td>
            <td>{{ $slot.Date }}</td>
            <td>{{ $slot.Vaccine }}</td>
            <td>{{ $slot.Age }}</td>
            <td>{{ $slot.FeeType }}</td>
        </tr>
        {{end}}
    </table>

</body>

</html>
`
