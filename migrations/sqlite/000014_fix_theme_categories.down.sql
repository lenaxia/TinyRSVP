-- Revert theme categories back to coarse values
UPDATE templates SET category = 'plain' WHERE name = 'Simple & Clean'        AND type = 'rsvp_page';
UPDATE templates SET category = 'card'  WHERE name = 'Wedding Elegance'      AND type = 'rsvp_page';
UPDATE templates SET category = 'card'  WHERE name = 'Birthday Celebration'  AND type = 'rsvp_page';
UPDATE templates SET category = 'card'  WHERE name = 'Corporate Professional' AND type = 'rsvp_page';
UPDATE templates SET category = 'card'  WHERE name = 'Holiday Festive'       AND type = 'rsvp_page';
UPDATE templates SET category = 'card'  WHERE name = 'Garden Party'          AND type = 'rsvp_page';
UPDATE templates SET category = 'card'  WHERE name = 'Modern Minimalist'     AND type = 'rsvp_page';
