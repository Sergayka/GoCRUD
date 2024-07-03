document.addEventListener('DOMContentLoaded', function() {
    function fetchPersons() {
        fetch('/persons')
            .then(response => response.json())
            .then(data => {
                let personTableBody = document.querySelector('#personTable tbody');
                personTableBody.innerHTML = '';

                data.forEach(person => {
                    let row = document.createElement('tr');
                    row.innerHTML = `
                        <td>${person.first_name}</td>
                        <td>${person.last_name}</td>
                        <td>${person.city}</td>
                        <td>
                            <button class="edit-btn" data-id="${person.id}">✏️</button>
                            <button class="delete-btn" data-id="${person.id}">❌</button>
                        </td>
                    `;
                    personTableBody.appendChild(row);
                });

                //TODO: Добавить обработчик события удаления
                document.querySelectorAll('.delete-btn').forEach(button => {
                    button.addEventListener('click', function() {
                        const personId = this.getAttribute('data-id');
                        deletePerson(personId);
                    });
                });

                //TODO: Добавить обработчик события редактирования
                document.querySelectorAll('.edit-btn').forEach(button => {
                    button.addEventListener('click', function() {
                        const personId = this.getAttribute('data-id');
                        editPerson(personId);
                    });
                });
            });
    }

    function deletePerson(id) {
        fetch(`/persons/${id}`, {
            method: 'DELETE'
        })
            .then(response => response.json())
            .then(data => {
                fetchPersons();
            })
            .catch(error => console.error('Error:', error));
    }

    function editPerson(id) {
        fetch(`/persons/${id}`)
            .then(response => response.json())
            .then(person => {
                // console.log(person)
                document.getElementById('editPersonId').value = person.id;
                document.getElementById('editFirstName').value = person.first_name;
                document.getElementById('editLastName').value = person.last_name;
                document.getElementById('editCity').value = person.city;
                document.getElementById('editPersonForm').style.display = 'block';
            })
            .catch(error => console.error('Error:', error));
    }

    fetchPersons(); // Загрузить и отобразить текущих пользователей при загрузке страницы

    const form = document.getElementById('personForm');
    form.addEventListener('submit', function(event) {
        event.preventDefault();

        const formData = new FormData(form);
        const person = {
            first_name: formData.get('firstName'),
            last_name: formData.get('lastName'),
            city: formData.get('city')
        };

        fetch('/persons', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(person)
        })
            .then(response => response.json())
            .then(data => {
                fetchPersons(); // Обновить таблицу после добавления нового пользователя
            })
            .catch(error => console.error('Error:', error));
    });

    const editForm = document.getElementById('editPersonForm');
    editForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const personId = document.getElementById('editPersonId').value;
        const formData = new FormData(editForm);
        const person = {
            first_name: formData.get('firstName'),
            last_name: formData.get('lastName'),
            city: formData.get('city')
        };

        fetch(`/persons/${personId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(person)
        })
            .then(response => response.json())
            .then(data => {
                fetchPersons(); // Обновить таблицу после редактирования пользователя
                document.getElementById('editPersonForm').style.display = 'none';
            })
            .catch(error => console.error('Error:', error));
    });
});
