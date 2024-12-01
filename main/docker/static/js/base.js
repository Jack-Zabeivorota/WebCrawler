const domain = 'http://localhost:8000'

// Alert

function setCloseAlertEvent() {
    const alert = document.getElementById('alert')
    const closeBtn = alert.querySelector('.close')

    closeBtn.addEventListener('click', () =>
        alert.classList.remove('show', 'error', 'info', 'success')
    )
}

function showAlert(message, status) {
    const alert    = document.getElementById('alert')
    const msgBlock = alert.querySelector('.message')

    msgBlock.innerHTML = message;
    alert.classList.add('show', status)
}

// Helpers

function setRequestResults(data) {
    let urlsWithWords = 0

    const urlsData = [`
        <tr>
            <th>URL</th>
            <th>Words</th>
            <th>Status</th>
        </tr>`
    ];
    
    for (const urlData of data.urls) {
        let withoutWords = true

        if (urlData.finded_words != '') {
            urlsWithWords++
            withoutWords = false
        }

        urlsData.push(`
            <tr${withoutWords ? ' class="without-words"' : ''}>
                <td>${urlData.url}</td>
                <td>${urlData.finded_words.join(', ')}</td>
                <td class="${urlData.status === 'Success' ? 'success' : 'fail'}">
                    ${urlData.status}
                </td>
            </tr>`
        )
    }

    const infoBlock = document.querySelector('.info-block')
    const labels    = infoBlock.querySelector('.labels')
    const results   = infoBlock.querySelector('.results')
    const toogler   = infoBlock.querySelector('.show-all-urls')

    infoBlock.classList.remove('hide')
    toogler.checked = false

    labels.innerHTML = `
        <tr>
            <td>Start url:</td>
            <td>${data.start_url}</td>
        </tr>
        <tr>
            <td>Search words:</td>
            <td>${data.words.join(', ')}</td>
        </tr>
        <tr>
            <td>Same domain only:</td>
            <td>${data.same_domain_only ? 'Yes' : 'No'}</td>
        </tr>
        <tr>
            <td>All finded URLs:</td>
            <td>${data.urls.length}</td>
        </tr>
        <tr>
            <td>URLs with words:</td>
            <td>${urlsWithWords}</td>
        </tr>`
    
    results.innerHTML = urlsData.join('\n')
}

function hideRequestResults() {
    const infoBlock = document.querySelector('.info-block')
    infoBlock.classList.add('hide')
}

// Events

function setMenuEvents() {
    const menuItems = document.querySelectorAll('header > div')
    const sections  = document.querySelectorAll('section')

    for (let i = 0; i < menuItems.length; i++) {
        menuItems[i].addEventListener('click', () => {
            for (const item of menuItems)
                item.classList.remove('active')

            for (const sect of sections)
                sect.classList.remove('active')

            menuItems[i].classList.add('active')
            sections[i].classList.add('active')
        })
    }
}

function setShowAllURLsEvent() {
    const toogler = document.querySelector('.info-block .show-all-urls input')

    toogler.addEventListener('change', () => {
        const withoutWordsUrls = document.querySelectorAll('.info-block .results tr.without-words')

        if (toogler.checked) {
            for (const cell of withoutWordsUrls)
                cell.classList.add('active')
        } else {
            for (const cell of withoutWordsUrls)
                cell.classList.remove('active')
        }
    })
}

// Submits

function setCreateRequestEvent() {
    const submit = document.querySelector('#create-request .field-block .submit')

    submit.addEventListener('click', async () => {
        const url        = document.querySelector('#create-request .field-block input').value
        const words      = document.querySelector('#create-request .words-block input').value
        const sameDomain = document.querySelector('#create-request .same-domain-block input').checked

        if (url === '' || words === '') {
            showAlert('Field is empty', 'error')
            return
        }
        let data

        try {
            const responce = await fetch(domain + '/request', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    url: url,
                    same_domain_only: sameDomain,
                    words: words.replace(/\s+/g, '').split(',').filter(w => w != ''),
                })
            })
            data = await responce.json()

            if (!responce.ok) {
                showAlert(data.message, 'error')
                return
            }
        } catch {
            showAlert('Server error, try later', 'error')
            return
        }

        showAlert(`${data.message}<br>Your request ID: ${data.request_id}`, 'success')
    })
}

function setCheckRequestEvent() {
    const submit = document.querySelector('#check-request .field-block .submit')
    
    submit.addEventListener('click', async () => {
        const requestID = document.querySelector('#check-request .field-block input').value

        if (requestID === '') {
            hideRequestResults()
            showAlert('Enter request ID', 'error')
            return
        }
        let data

        try {
            const responce = await fetch(`${domain}/request?ID=${requestID}`)
            data = await responce.json()

            if (!responce.ok) {
                hideRequestResults()
                showAlert(data.message, 'error')
                return
            }
        } catch {
            hideRequestResults()
            showAlert('Server error, try later', 'error')
            return
        }
        
        if (data.message !== undefined) {
            showAlert(data.message, 'info')
            return
        }

        setRequestResults(data)
    })
}

function setDeleteRequestEvent() {
    const submit = document.querySelector('#delete-request .field-block .submit')
    
    submit.addEventListener('click', async () => {
        const requestID = document.querySelector('#delete-request .field-block input').value

        if (requestID === '') {
            showAlert('Enter request ID', 'error')
            return
        }
        let data

        try {
            const responce = await fetch(`${domain}/request?ID=${requestID}`, {
                method: 'DELETE',
            })
            data = await responce.json()

            if (!responce.ok) {
                showAlert(data.message, 'error')
                return
            }
        } catch {
            showAlert('Server error, try later', 'error')
            return
        }
        
        showAlert(data.message, 'success')
    })
}


document.addEventListener('DOMContentLoaded', () => {
    setMenuEvents()
    setCloseAlertEvent()
    setShowAllURLsEvent()

    setCreateRequestEvent()
    setCheckRequestEvent()
    setDeleteRequestEvent()
})