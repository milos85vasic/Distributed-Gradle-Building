// Main JavaScript functionality
document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl)
    });

    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Copy code buttons
    function addCopyButtons() {
        const codeBlocks = document.querySelectorAll('pre code');
        codeBlocks.forEach(function(block) {
            // Only add copy button if it doesn't already exist
            const wrapper = block.parentElement;
            if (!wrapper.querySelector('.copy-button')) {
                const button = document.createElement('button');
                button.className = 'copy-button btn btn-sm btn-outline-secondary position-absolute';
                button.style.top = '10px';
                button.style.right = '10px';
                button.innerHTML = '<i class="fas fa-copy"></i> Copy';
                button.onclick = function() {
                    navigator.clipboard.writeText(block.textContent).then(function() {
                        button.innerHTML = '<i class="fas fa-check"></i> Copied!';
                        setTimeout(function() {
                            button.innerHTML = '<i class="fas fa-copy"></i> Copy';
                        }, 2000);
                    });
                };
                wrapper.style.position = 'relative';
                wrapper.appendChild(button);
            }
        });
    }

    // Add copy buttons to code blocks
    addCopyButtons();

    // Search functionality
    const searchInput = document.getElementById('search-input');
    const searchResults = document.getElementById('search-results');
    
    if (searchInput && searchResults) {
        searchInput.addEventListener('input', function() {
            const query = this.value.toLowerCase();
            if (query.length < 2) {
                searchResults.innerHTML = '';
                searchResults.classList.add('d-none');
                return;
            }

            // Simple search implementation (would need backend for real search)
            const searchableContent = document.querySelectorAll('.content h1, .content h2, .content h3, .content p');
            const results = [];
            
            searchableContent.forEach(function(element) {
                const text = element.textContent.toLowerCase();
                if (text.includes(query)) {
                    results.push({
                        title: element.textContent.substring(0, 60),
                        element: element
                    });
                }
            });

            if (results.length > 0) {
                searchResults.innerHTML = results.map(function(result) {
                    return `
                        <div class="list-group-item">
                            <h6>${result.title}</h6>
                        </div>
                    `;
                }).join('');
                searchResults.classList.remove('d-none');
            } else {
                searchResults.innerHTML = '<div class="list-group-item">No results found</div>';
                searchResults.classList.remove('d-none');
            }
        });

        // Hide search results when clicking outside
        document.addEventListener('click', function(e) {
            if (!searchInput.contains(e.target) && !searchResults.contains(e.target)) {
                searchResults.classList.add('d-none');
            }
        });
    }

    // Table of contents highlighting
    const tocLinks = document.querySelectorAll('.table-of-contents a');
    const sections = document.querySelectorAll('h2[id], h3[id], h4[id]');

    function highlightCurrentSection() {
        const scrollPosition = window.scrollY + 100;
        
        sections.forEach(section => {
            const sectionTop = section.offsetTop;
            const sectionHeight = section.offsetHeight;
            const sectionId = section.getAttribute('id');
            
            if (scrollPosition >= sectionTop && scrollPosition < sectionTop + sectionHeight) {
                tocLinks.forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('href') === '#' + sectionId) {
                        link.classList.add('active');
                    }
                });
            }
        });
    }

    if (tocLinks.length > 0) {
        window.addEventListener('scroll', highlightCurrentSection);
        highlightCurrentSection(); // Initialize
    }

    // Progress indicator for documentation
    const progressBar = document.getElementById('reading-progress');
    if (progressBar) {
        function updateProgressBar() {
            const scrollTop = window.scrollY;
            const documentHeight = document.documentElement.scrollHeight - window.innerHeight;
            const scrollPercent = (scrollTop / documentHeight) * 100;
            progressBar.style.width = scrollPercent + '%';
        }
        
        window.addEventListener('scroll', updateProgressBar);
        updateProgressBar(); // Initialize
    }

    // Theme toggle functionality
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        // Check for saved theme preference
        const savedTheme = localStorage.getItem('theme') || 'light';
        document.body.classList.add('theme-' + savedTheme);
        
        themeToggle.addEventListener('click', function() {
            const currentTheme = document.body.classList.contains('theme-dark') ? 'dark' : 'light';
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            
            document.body.classList.remove('theme-' + currentTheme);
            document.body.classList.add('theme-' + newTheme);
            localStorage.setItem('theme', newTheme);
            
            // Update button icon
            this.innerHTML = newTheme === 'dark' ? 
                '<i class="fas fa-sun"></i>' : 
                '<i class="fas fa-moon"></i>';
        });
    }

    // Auto-hide navbar on scroll down, show on scroll up
    let lastScrollTop = 0;
    const navbar = document.querySelector('.navbar');
    
    window.addEventListener('scroll', function() {
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
        
        if (scrollTop > lastScrollTop && scrollTop > 100) {
            navbar.style.transform = 'translateY(-100%)';
        } else {
            navbar.style.transform = 'translateY(0)';
        }
        
        lastScrollTop = scrollTop;
    });

    // Add smooth transitions to navbar
    navbar.style.transition = 'transform 0.3s ease';
});

// Utility functions
function scrollToElement(elementId) {
    const element = document.getElementById(elementId);
    if (element) {
        element.scrollIntoView({ behavior: 'smooth' });
    }
}

function toggleElement(elementId) {
    const element = document.getElementById(elementId);
    if (element) {
        element.classList.toggle('d-none');
    }
}

// Error handling
window.addEventListener('error', function(e) {
    console.error('JavaScript error:', e.error);
});

// Performance monitoring
if (window.performance) {
    window.addEventListener('load', function() {
        const timing = window.performance.timing;
        const loadTime = timing.loadEventEnd - timing.navigationStart;
        console.log('Page load time:', loadTime + 'ms');
    });
}