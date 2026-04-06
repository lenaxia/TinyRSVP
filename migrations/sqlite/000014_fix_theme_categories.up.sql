-- Fix theme categories: each theme now has a unique slug-based category
-- that directly matches the CSS filename in static/css/themes/
UPDATE templates SET category = 'wedding-elegance'     WHERE name = 'Wedding Elegance'      AND type = 'rsvp_page';
UPDATE templates SET category = 'birthday-celebration' WHERE name = 'Birthday Celebration'  AND type = 'rsvp_page';
UPDATE templates SET category = 'corporate-professional' WHERE name = 'Corporate Professional' AND type = 'rsvp_page';
UPDATE templates SET category = 'holiday-festive'      WHERE name = 'Holiday Festive'       AND type = 'rsvp_page';
UPDATE templates SET category = 'garden-party'         WHERE name = 'Garden Party'          AND type = 'rsvp_page';
UPDATE templates SET category = 'modern-minimalist'    WHERE name = 'Modern Minimalist'     AND type = 'rsvp_page';
UPDATE templates SET category = 'plain-text'           WHERE name = 'Simple & Clean'        AND type = 'rsvp_page';
